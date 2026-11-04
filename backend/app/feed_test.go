package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const feedToken = "abcdefghijklmnopqrstuvwxyz012345"

// feedApp has one person whose feed secret is feedToken, and three entries.
func feedApp(t testing.TB) (*tests.TestApp, *core.Record, string) {
	a := newApp(t)
	u, token := addUser(t, a, "ada@example.com", "member")
	u.Set("feed_token", feedToken)
	if err := a.Save(u); err != nil {
		t.Fatal(err)
	}
	addEntry(t, a, map[string]any{
		"title": "Crete, with friends", "kind": "trip", "all_day": true,
		"start": "2026-12-12 00:00:00.000Z", "end": "2026-12-18 00:00:00.000Z",
		"checklist": `[{"text":"Passports","done":true}]`,
	})
	addEntry(t, a, map[string]any{
		"title": "Piano", "kind": "event", "repeat": "weekly",
		"start": "2026-12-02 16:30:00.000Z", "end": "2026-12-02 17:30:00.000Z",
	})
	addEntry(t, a, map[string]any{
		"title": "Luna", "kind": "petsit", "all_day": true, "sitter": "Nikos", "sitter_status": "booked",
		"start": "2026-12-12 00:00:00.000Z",
	})
	return a, u, token
}

func TestFeed(t *testing.T) {
	t.Setenv("CALIO_FEED_URL", "https://calio-feed.example.ts.net/")
	scenarios := []struct {
		name, method, url string
		auth              bool
		status            int
		content, not      []string
		after             func(testing.TB, *tests.TestApp, *http.Response)
	}{
		{
			name: "the feed lists everything", method: http.MethodGet,
			url: "/api/app/feed/" + feedToken + ".ics", status: http.StatusOK,
			content: []string{
				"BEGIN:VCALENDAR", "X-WR-CALNAME:Calio",
				"SUMMARY:✈️ Crete\\, with friends",
				"DTSTART;VALUE=DATE:20261212", "DTEND;VALUE=DATE:20261219", // the day after
				"DESCRIPTION:☑ Passports",
				"SUMMARY:Piano", "DTSTART:20261202T163000Z", "RRULE:FREQ=WEEKLY",
				"SUMMARY:🐾 Luna (Nikos is booked)",
			},
			after: func(t testing.TB, _ *tests.TestApp, res *http.Response) {
				if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/calendar") {
					t.Errorf("Content-Type = %q", ct)
				}
			},
		},
		{
			name: "a wrong secret finds nothing", method: http.MethodGet,
			url: "/api/app/feed/" + strings.Repeat("x", 32) + ".ics", status: http.StatusNotFound,
			not: []string{"VCALENDAR"},
		},
		{
			name: "the feed link needs a signed-in person", method: http.MethodGet,
			url: "/api/app/feed-link", status: http.StatusUnauthorized,
			content: []string{`"status":401`},
		},
		{
			name: "the feed link uses the public address", method: http.MethodGet,
			url: "/api/app/feed-link", auth: true, status: http.StatusOK,
			content: []string{`"url":"https://calio-feed.example.ts.net/api/app/feed/` + feedToken + `.ics"`, `"public":true`},
		},
		{
			name: "a new link stops the old one working", method: http.MethodPost,
			url: "/api/app/feed-link/reset", auth: true, status: http.StatusOK,
			content: []string{`"url":"https://calio-feed.example.ts.net/api/app/feed/`},
			not:     []string{feedToken},
			after: func(t testing.TB, a *tests.TestApp, _ *http.Response) {
				if _, err := a.FindFirstRecordByData("users", "feed_token", feedToken); err == nil {
					t.Error("the old secret still works")
				}
			},
		},
		{
			name: "the secret never shows in the API", method: http.MethodGet,
			url: "/api/collections/users/records", auth: true, status: http.StatusOK,
			not: []string{feedToken, "feed_token"},
		},
	}
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			a, _, token := feedApp(t)
			headers := map[string]string{}
			if sc.auth {
				headers["Authorization"] = token
			}
			(&tests.ApiScenario{
				Method:             sc.method,
				URL:                sc.url,
				Headers:            headers,
				ExpectedStatus:     sc.status,
				ExpectedContent:    sc.content,
				NotExpectedContent: sc.not,
				TestAppFactory:     func(testing.TB) *tests.TestApp { return a },
				AfterTestFunc:      sc.after,
			}).Test(t)
		})
	}
}
