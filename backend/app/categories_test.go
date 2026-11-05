package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func category(t testing.TB, a core.App, name string) *core.Record {
	c, err := a.FindFirstRecordByData("categories", "name", name)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCategories(t *testing.T) {
	t.Run("a household starts with four, babysitting switched off", func(t *testing.T) {
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		if category(t, a, "Babysitting").GetBool("enabled") {
			t.Error("babysitting starts switched on")
		}
		(&tests.ApiScenario{
			Method:         http.MethodGet,
			URL:            "/api/collections/categories/records?sort=sort",
			Headers:        map[string]string{"Authorization": token},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				`"totalItems":4`, `"name":"Event"`, `"name":"Pet sitting"`, `"name":"Trip"`,
				`"name":"Babysitting"`,
			},
			TestAppFactory: func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("a member adds one with a color", func(t *testing.T) {
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/categories/records",
			Body:            strings.NewReader(`{"name":"House sitting","color":"emerald","icon":"home","kind":"petsit","enabled":true}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"color":"emerald"`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("an entry's kind follows its category", func(t *testing.T) {
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		baby := category(t, a, "Babysitting")
		(&tests.ApiScenario{
			Method: http.MethodPost,
			URL:    "/api/collections/entries/records",
			Body: strings.NewReader(`{"title":"Eva and Leo","kind":"event","category":"` + baby.Id +
				`","start":"2026-12-12 00:00:00.000Z","all_day":true}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"kind":"petsit"`, `"sitter_status":"needed"`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("deleting a category keeps its entries", func(t *testing.T) {
		a := newApp(t)
		trip := category(t, a, "Trip")
		e := addEntry(t, a, map[string]any{
			"title": "Crete", "kind": "trip", "category": trip.Id, "all_day": true,
			"start": "2026-12-12 00:00:00.000Z",
		})
		if err := a.Delete(trip); err != nil {
			t.Fatal(err)
		}
		e, err := a.FindRecordById("entries", e.Id)
		if err != nil {
			t.Fatal(err)
		}
		if e.GetString("category") != "" || e.GetString("kind") != "trip" {
			t.Errorf("category = %q, kind = %q", e.GetString("category"), e.GetString("kind"))
		}
	})
}
