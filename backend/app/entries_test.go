package app_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"

	"github.com/selfhostit/calio/app"
)

func addEntry(t testing.TB, a core.App, fields map[string]any) *core.Record {
	return addRecord(t, a, "entries", fields)
}

func addRecord(t testing.TB, a core.App, collection string, fields map[string]any) *core.Record {
	c, err := a.FindCollectionByNameOrId(collection)
	if err != nil {
		t.Fatal(err)
	}
	r := core.NewRecord(c)
	r.Load(fields)
	if err := a.Save(r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestEntries(t *testing.T) {
	scenarios := []struct {
		name    string
		role    string
		body    string
		status  int
		content []string
	}{
		{
			name:    "a member adds a trip",
			role:    "member",
			body:    `{"title":"Crete","kind":"trip","start":"2026-12-12 00:00:00.000Z","end":"2026-12-18 00:00:00.000Z","all_day":true}`,
			status:  http.StatusOK,
			content: []string{`"title":"Crete"`, `"kind":"trip"`},
		},
		{
			name:    "pet sitting starts out needing a sitter",
			role:    "member",
			body:    `{"title":"Luna","kind":"petsit","start":"2026-12-12 00:00:00.000Z","all_day":true}`,
			status:  http.StatusOK,
			content: []string{`"sitter_status":"needed"`},
		},
		{
			name:    "the end can't be before the start",
			role:    "member",
			body:    `{"title":"Oops","kind":"event","start":"2026-12-12 10:00:00.000Z","end":"2026-12-12 09:00:00.000Z"}`,
			status:  http.StatusBadRequest,
			content: []string{`"end"`},
		},
		{
			name:    "an entry needs a kind",
			role:    "member",
			body:    `{"title":"Party","start":"2026-12-12 10:00:00.000Z"}`,
			status:  http.StatusBadRequest,
			content: []string{`"kind"`},
		},
	}
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			a := newApp(t)
			u, token := addUser(t, a, "someone@example.com", sc.role)
			s := tests.ApiScenario{
				Method:          http.MethodPost,
				URL:             "/api/collections/entries/records",
				Body:            strings.NewReader(sc.body),
				Headers:         map[string]string{"Authorization": token},
				ExpectedStatus:  sc.status,
				ExpectedContent: sc.content,
				TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
			}
			if sc.status == http.StatusOK {
				s.ExpectedContent = append(s.ExpectedContent, `"created_by":"`+u.Id+`"`)
			}
			s.Test(t)
		})
	}
}

func TestOnlyTheAuthorOrOwnerDeletes(t *testing.T) {
	for _, c := range []struct {
		name   string
		as     string
		status int
	}{
		{"another member can't", "other", http.StatusNotFound},
		{"the author can", "author", http.StatusNoContent},
		{"the owner can", "owner", http.StatusNoContent},
	} {
		t.Run(c.name, func(t *testing.T) {
			a := newApp(t)
			author, authorToken := addUser(t, a, "author@example.com", "member")
			_, otherToken := addUser(t, a, "other@example.com", "member")
			_, ownerToken := addUser(t, a, "owner@example.com", "owner")
			token := map[string]string{"author": authorToken, "other": otherToken, "owner": ownerToken}[c.as]
			entry := addEntry(t, a, map[string]any{
				"title": "Dinner", "kind": "event", "start": "2026-12-12 18:00:00.000Z", "created_by": author.Id,
			})
			s := tests.ApiScenario{
				Method:         http.MethodDelete,
				URL:            "/api/collections/entries/records/" + entry.Id,
				Headers:        map[string]string{"Authorization": token},
				ExpectedStatus: c.status,
				TestAppFactory: func(testing.TB) *tests.TestApp { return a },
			}
			if c.status == http.StatusNotFound {
				s.ExpectedContent = []string{`"status":404`}
			}
			s.Test(t)
		})
	}
}

func TestSignedOutVisitorsSeeNothing(t *testing.T) {
	a := newApp(t)
	addEntry(t, a, map[string]any{"title": "Secret", "kind": "event", "start": "2026-12-12 18:00:00.000Z"})
	(&tests.ApiScenario{
		Method:             http.MethodGet,
		URL:                "/api/collections/entries/records",
		ExpectedStatus:     http.StatusOK,
		ExpectedContent:    []string{`"totalItems":0`},
		NotExpectedContent: []string{"Secret"},
		TestAppFactory:     func(testing.TB) *tests.TestApp { return a },
	}).Test(t)
}

func TestMorningDigest(t *testing.T) {
	a := newApp(t)
	now := time.Date(2026, 12, 12, 7, 30, 0, 0, time.Local)
	addEntry(t, a, map[string]any{
		"title": "Crete", "kind": "trip", "all_day": true,
		"start": "2026-12-10 00:00:00.000Z", "end": "2026-12-14 00:00:00.000Z",
	})
	addEntry(t, a, map[string]any{
		"title": "Luna", "kind": "petsit", "all_day": true,
		"start": "2026-12-15 00:00:00.000Z", "end": "2026-12-17 00:00:00.000Z",
	})
	addEntry(t, a, map[string]any{
		"title": "Last year", "kind": "event", "all_day": true, "start": "2025-12-12 00:00:00.000Z",
	})
	addEntry(t, a, map[string]any{
		"title": "Booked", "kind": "petsit", "all_day": true, "sitter_status": "booked",
		"start": "2026-12-14 00:00:00.000Z",
	})

	msg, ok, err := app.MorningDigest(a, now)
	if err != nil || !ok {
		t.Fatalf("ok = %v, err = %v", ok, err)
	}
	if msg.Title != "1 on today" {
		t.Errorf("title = %q", msg.Title)
	}
	want := "Today: Crete. Still needs a sitter: Luna (Tue 15 – Thu 17 Dec)"
	if msg.Body != want {
		t.Errorf("body = %q, want %q", msg.Body, want)
	}

	if _, ok, _ := app.MorningDigest(a, now.AddDate(0, 1, 0)); ok {
		t.Error("a quiet day still sent a digest")
	}
}

func TestMorningDigestRepeats(t *testing.T) {
	a := newApp(t)
	addEntry(t, a, map[string]any{
		"title": "Piano", "kind": "event", "repeat": "weekly",
		"start": time.Date(2026, 1, 7, 17, 0, 0, 0, time.Local), // a Wednesday
	})
	addEntry(t, a, map[string]any{
		"title": "Rent", "kind": "event", "all_day": true, "repeat": "monthly",
		"start": "2026-01-31 00:00:00.000Z", // skips months without a 31st
	})
	addEntry(t, a, map[string]any{
		"title": "Gym", "kind": "event", "all_day": true, "repeat": "daily",
		"start": "2026-01-01 00:00:00.000Z", "repeat_until": "2026-01-31 00:00:00.000Z",
	})

	for _, c := range []struct {
		day  time.Time
		want string
	}{
		{time.Date(2026, 12, 2, 7, 30, 0, 0, time.Local), "Today: Piano"},
		{time.Date(2026, 12, 30, 7, 30, 0, 0, time.Local), "Today: Piano"},
		{time.Date(2026, 12, 31, 7, 30, 0, 0, time.Local), "Today: Rent"},
		{time.Date(2026, 1, 20, 7, 30, 0, 0, time.Local), "Today: Gym"},
	} {
		msg, _, err := app.MorningDigest(a, c.day)
		if err != nil {
			t.Fatal(err)
		}
		if msg.Body != c.want {
			t.Errorf("%s: body = %q, want %q", c.day.Format(time.DateOnly), msg.Body, c.want)
		}
	}
	if _, ok, _ := app.MorningDigest(a, time.Date(2026, 11, 30, 7, 30, 0, 0, time.Local)); ok {
		t.Error("30 November has no rent, piano or gym")
	}
}
