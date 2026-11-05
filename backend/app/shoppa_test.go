package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/pocketbase/pocketbase/tests"

	"github.com/selfhostit/calio/app"
)

// fakeShoppa answers like Shoppa does, and keeps what was added.
func fakeShoppa(t *testing.T) *[]map[string]any {
	var mu sync.Mutex
	added := &[]map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/collections/workers/auth-with-password":
			var in map[string]string
			_ = json.NewDecoder(r.Body).Decode(&in)
			if in["identity"] != "calio" || in["password"] != "secret" {
				http.Error(w, "no", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"token":"tok"}`))
		case r.Header.Get("Authorization") != "tok":
			http.Error(w, "no", http.StatusUnauthorized)
		case r.URL.Path == "/api/collections/lists/records":
			_, _ = w.Write([]byte(`{"items":[{"id":"groceries","name":"Groceries"},{"id":"bbq","name":"BBQ"}]}`))
		case r.URL.Path == "/api/app/items":
			var in map[string]any
			_ = json.NewDecoder(r.Body).Decode(&in)
			mu.Lock()
			*added = append(*added, in)
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("SHOPPA_URL", srv.URL+"/")
	t.Setenv("SHOPPA_WORKER", "calio")
	t.Setenv("SHOPPA_PASSWORD", "secret")
	return added
}

func TestShoppa(t *testing.T) {
	t.Run("off until it is set up", func(t *testing.T) {
		t.Setenv("SHOPPA_URL", "")
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		(&tests.ApiScenario{
			Method: http.MethodGet, URL: "/api/app/shoppa",
			Headers:        map[string]string{"Authorization": token},
			ExpectedStatus: http.StatusOK, ExpectedContent: []string{`"enabled":false`},
			TestAppFactory: func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("sends what is still to buy", func(t *testing.T) {
		added := fakeShoppa(t)
		t.Setenv("SHOPPA_LIST", "bbq")
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		e := addEntry(t, a, map[string]any{
			"title": "BBQ", "kind": "event", "start": "2026-12-12 18:00:00.000Z",
			"shopping": `[{"text":"2 worst","done":false},{"text":"Bread","done":true},{"text":"Houtskool","done":false}]`,
		})
		(&tests.ApiScenario{
			Method: http.MethodPost, URL: "/api/app/shoppa/send/" + e.Id,
			Headers:        map[string]string{"Authorization": token},
			ExpectedStatus: http.StatusOK, ExpectedContent: []string{`"sent":2`, `"list":"BBQ"`},
			TestAppFactory: func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
		got, _ := json.Marshal(*added)
		want := `[{"list":"bbq","name":"worst","quantity":"2×"},{"list":"bbq","name":"Houtskool"}]`
		if string(got) != want {
			t.Errorf("added %s, want %s", got, want)
		}
	})

	t.Run("a wrong password is Shoppa's problem, said plainly", func(t *testing.T) {
		fakeShoppa(t)
		t.Setenv("SHOPPA_PASSWORD", "wrong")
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		e := addEntry(t, a, map[string]any{
			"title": "BBQ", "kind": "event", "start": "2026-12-12 18:00:00.000Z",
			"shopping": `[{"text":"Bread","done":false}]`,
		})
		(&tests.ApiScenario{
			Method: http.MethodPost, URL: "/api/app/shoppa/send/" + e.Id,
			Headers:        map[string]string{"Authorization": token},
			ExpectedStatus: http.StatusBadGateway, ExpectedContent: []string{"Couldn't reach Shoppa"},
			TestAppFactory: func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})
}

func TestParseItem(t *testing.T) {
	for in, want := range map[string]string{
		"2 melk": "melk|2×", "melk x2": "melk|2×", "2xmelk": "melk|2×", "melk 3×": "melk|3×",
		"7up": "7up|", "Cola 1,5 L": "Cola 1,5 L|",
	} {
		name, q := app.ParseItem(in)
		if got := name + "|" + q; got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
