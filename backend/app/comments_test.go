package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/tests"
)

func TestComments(t *testing.T) {
	t.Run("the author is whoever writes it", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		other, _ := addUser(t, a, "bo@example.com", "member")
		e := addEntry(t, a, map[string]any{"title": "Dinner", "kind": "event", "start": "2026-12-12 18:00:00.000Z"})
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/comments/records",
			Body:            strings.NewReader(`{"entry":"` + e.Id + `","text":"I'll bring wine","author":"` + other.Id + `"}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"author":"` + u.Id + `"`, `"text":"I'll bring wine"`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("only the author edits it", func(t *testing.T) {
		a := newApp(t)
		u, _ := addUser(t, a, "ada@example.com", "member")
		_, otherToken := addUser(t, a, "bo@example.com", "member")
		e := addEntry(t, a, map[string]any{"title": "Dinner", "kind": "event", "start": "2026-12-12 18:00:00.000Z"})
		comments, _ := a.FindCollectionByNameOrId("comments")
		c := addRecord(t, a, comments.Name, map[string]any{"entry": e.Id, "text": "Hi", "author": u.Id})
		(&tests.ApiScenario{
			Method:          http.MethodPatch,
			URL:             "/api/collections/comments/records/" + c.Id,
			Body:            strings.NewReader(`{"text":"Changed"}`),
			Headers:         map[string]string{"Authorization": otherToken},
			ExpectedStatus:  http.StatusNotFound,
			ExpectedContent: []string{`"status":404`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("comments go with their entry", func(t *testing.T) {
		a := newApp(t)
		u, _ := addUser(t, a, "ada@example.com", "member")
		e := addEntry(t, a, map[string]any{"title": "Dinner", "kind": "event", "start": "2026-12-12 18:00:00.000Z"})
		addRecord(t, a, "comments", map[string]any{"entry": e.Id, "text": "Hi", "author": u.Id})
		if err := a.Delete(e); err != nil {
			t.Fatal(err)
		}
		if n, _ := a.CountRecords("comments"); n != 0 {
			t.Errorf("%d comments left", n)
		}
	})
}
