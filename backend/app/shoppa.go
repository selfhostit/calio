package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// Shoppa (github.com/selfhostit/shoppa) is the household's shopping list.
// When it is set up, an entry's shopping list can be sent to it. Calio signs
// in to Shoppa as a worker account made in Shoppa's admin page:
//
//	SHOPPA_URL       Shoppa's address, like http://shoppa:8090
//	SHOPPA_WORKER    the worker's name
//	SHOPPA_PASSWORD  its password
//	SHOPPA_LIST      the list to add to (optional; the first list otherwise)
type shoppa struct {
	url, worker, password, list string
}

func shoppaFromEnv() (shoppa, bool) {
	s := shoppa{
		url:      strings.TrimRight(os.Getenv("SHOPPA_URL"), "/"),
		worker:   os.Getenv("SHOPPA_WORKER"),
		password: os.Getenv("SHOPPA_PASSWORD"),
		list:     strings.TrimSpace(os.Getenv("SHOPPA_LIST")),
	}
	return s, s.url != "" && s.worker != "" && s.password != ""
}

var shoppaClient = &http.Client{Timeout: 15 * time.Second}

func registerShoppaRoutes(se *core.ServeEvent) {
	g := se.Router.Group("/api/app/shoppa").Bind(apis.RequireAuth("users"))
	g.GET("", func(e *core.RequestEvent) error {
		_, ok := shoppaFromEnv()
		return e.JSON(http.StatusOK, map[string]any{"enabled": ok})
	})
	g.POST("/send/{entry}", sendToShoppa)
}

// sendToShoppa adds what is still to buy on an entry's shopping list to
// Shoppa. Shoppa keeps one item per thing, so sending twice adds nothing new.
func sendToShoppa(e *core.RequestEvent) error {
	s, ok := shoppaFromEnv()
	if !ok {
		return e.NotFoundError("Shoppa isn't set up.", nil)
	}
	entry, err := e.App.FindRecordById("entries", e.Request.PathValue("entry"))
	if err != nil {
		return e.NotFoundError("That entry is gone.", err)
	}
	var items []struct {
		Text string `json:"text"`
		Done bool   `json:"done"`
	}
	_ = entry.UnmarshalJSONField("shopping", &items)
	var names []string
	for _, i := range items {
		if t := strings.TrimSpace(i.Text); t != "" && !i.Done {
			names = append(names, t)
		}
	}
	if len(names) == 0 {
		return e.BadRequestError("There's nothing left to buy on this list.", nil)
	}
	list, err := s.send(names)
	if err != nil {
		e.App.Logger().Error("sending to Shoppa", "error", err)
		return e.Error(http.StatusBadGateway, "Couldn't reach Shoppa. Check SHOPPA_URL and the worker's name and password.", nil)
	}
	return e.JSON(http.StatusOK, map[string]any{"sent": len(names), "list": list})
}

// send signs in, finds the list and adds each item; it gives the list's name.
func (s shoppa) send(names []string) (string, error) {
	var auth struct {
		Token string `json:"token"`
	}
	if err := s.call(http.MethodPost, "/api/collections/workers/auth-with-password", "",
		map[string]string{"identity": s.worker, "password": s.password}, &auth); err != nil {
		return "", fmt.Errorf("signing in: %w", err)
	}
	var lists struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	q := url.Values{"perPage": {"200"}, "sort": {"sort,created"}, "filter": {"archived = false"}}
	if err := s.call(http.MethodGet, "/api/collections/lists/records?"+q.Encode(), auth.Token, nil, &lists); err != nil {
		return "", fmt.Errorf("finding the list: %w", err)
	}
	if len(lists.Items) == 0 {
		return "", errors.New("shoppa has no lists")
	}
	list := lists.Items[0]
	for _, l := range lists.Items {
		if s.list != "" && strings.EqualFold(strings.TrimSpace(l.Name), s.list) {
			list = l
			break
		}
	}
	for _, n := range names {
		name, quantity := parseItem(n)
		body := map[string]any{"list": list.ID, "name": name}
		if quantity != "" {
			body["quantity"] = quantity
		}
		if err := s.call(http.MethodPost, "/api/app/items", auth.Token, body, nil); err != nil {
			return "", fmt.Errorf("adding %q: %w", n, err)
		}
	}
	return list.Name, nil
}

func (s shoppa) call(method, path, token string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, s.url+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	res, err := shoppaClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("%s %s: %s", method, path, res.Status)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(out)
}

// As Shoppa reads its add bar: "2 melk", "2x melk", "melk x2" and "melk 2×"
// are two of "melk".
var (
	quantityBefore      = regexp.MustCompile(`(?i)^(\d{1,3})\s*[x×*]?\s+(.+)$`)
	quantityBeforeTight = regexp.MustCompile(`(?i)^(\d{1,3})[x×*](.+)$`)
	quantityAfter       = regexp.MustCompile(`(?i)^(.+?)\s+(?:[x×*]\s*(\d{1,3})|(\d{1,3})\s*[x×*])$`)
)

func parseItem(text string) (name, quantity string) {
	s := strings.Join(strings.Fields(text), " ")
	m := quantityBefore.FindStringSubmatch(s)
	if m == nil {
		m = quantityBeforeTight.FindStringSubmatch(s)
	}
	if m != nil {
		return strings.TrimSpace(m[2]), m[1] + "×"
	}
	if m := quantityAfter.FindStringSubmatch(s); m != nil {
		n := m[2]
		if n == "" {
			n = m[3]
		}
		return strings.TrimSpace(m[1]), n + "×"
	}
	return s, ""
}
