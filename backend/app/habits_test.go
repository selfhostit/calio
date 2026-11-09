package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/tests"
)

func TestHabits(t *testing.T) {
	habit := func(t *testing.T, a *tests.TestApp, user string) string {
		return addRecord(t, a, "habits", map[string]any{"user": user, "title": "Vitamin D", "icon": "pill", "times": 1}).Id
	}

	t.Run("a habit is for whoever adds it", func(t *testing.T) {
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		other, _ := addUser(t, a, "bo@example.com", "member")
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/habits/records",
			Body:            strings.NewReader(`{"user":"` + other.Id + `","title":"Floss","icon":"tooth","times":1}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{`"status":400`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	for _, tc := range []struct {
		name   string
		owner  bool
		status int
		want   string
	}{
		{"its owner ticks a habit off", true, http.StatusOK, `"count":1`},
		{"nobody else ticks it off", false, http.StatusBadRequest, `"status":400`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := newApp(t)
			u, token := addUser(t, a, "ada@example.com", "member")
			_, otherToken := addUser(t, a, "bo@example.com", "member")
			if !tc.owner {
				token = otherToken
			}
			(&tests.ApiScenario{
				Method:          http.MethodPost,
				URL:             "/api/collections/habit_checks/records",
				Body:            strings.NewReader(`{"habit":"` + habit(t, a, u.Id) + `","day":"2026-11-09","count":1}`),
				Headers:         map[string]string{"Authorization": token},
				ExpectedStatus:  tc.status,
				ExpectedContent: []string{tc.want},
				TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
			}).Test(t)
		})
	}

	t.Run("one check per habit and day", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		h := habit(t, a, u.Id)
		addRecord(t, a, "habit_checks", map[string]any{"habit": h, "day": "2026-11-09", "count": 1})
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/habit_checks/records",
			Body:            strings.NewReader(`{"habit":"` + h + `","day":"2026-11-09","count":1}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{`"status":400`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("a check can't move to another day", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		c := addRecord(t, a, "habit_checks", map[string]any{"habit": habit(t, a, u.Id), "day": "2026-11-09", "count": 1})
		(&tests.ApiScenario{
			Method:          http.MethodPatch,
			URL:             "/api/collections/habit_checks/records/" + c.Id,
			Body:            strings.NewReader(`{"day":"2026-11-01"}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusNotFound,
			ExpectedContent: []string{`"status":404`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("habits and checks go with their person", func(t *testing.T) {
		a := newApp(t)
		u, _ := addUser(t, a, "ada@example.com", "member")
		addRecord(t, a, "habit_checks", map[string]any{"habit": habit(t, a, u.Id), "day": "2026-11-09", "count": 1})
		addRecord(t, a, "companions", map[string]any{"user": u.Id, "name": "Ember", "species": "red_panda"})
		if err := a.Delete(u); err != nil {
			t.Fatal(err)
		}
		for _, c := range []string{"habits", "habit_checks", "companions"} {
			if n, _ := a.CountRecords(c); n != 0 {
				t.Errorf("%d %s left", n, c)
			}
		}
	})
}

func TestCompanions(t *testing.T) {
	t.Run("one each, named by its person", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		addRecord(t, a, "companions", map[string]any{"user": u.Id, "name": "Ember", "species": "red_panda"})
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/companions/records",
			Body:            strings.NewReader(`{"user":"` + u.Id + `","name":"Second","species":"red_panda"}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{`"status":400`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("nobody else renames it", func(t *testing.T) {
		a := newApp(t)
		u, _ := addUser(t, a, "ada@example.com", "member")
		_, otherToken := addUser(t, a, "bo@example.com", "member")
		c := addRecord(t, a, "companions", map[string]any{"user": u.Id, "name": "Ember", "species": "red_panda"})
		(&tests.ApiScenario{
			Method:          http.MethodPatch,
			URL:             "/api/collections/companions/records/" + c.Id,
			Body:            strings.NewReader(`{"name":"Mine now"}`),
			Headers:         map[string]string{"Authorization": otherToken},
			ExpectedStatus:  http.StatusNotFound,
			ExpectedContent: []string{`"status":404`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})
}
