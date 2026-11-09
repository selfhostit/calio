package app_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"

	"github.com/selfhostit/calio/play/habits"
	"github.com/selfhostit/calio/play/rewards"
)

// addHabit saves a habit that was added some days ago.
func addHabit(t *testing.T, a core.App, user string, times int, daysAgo int) *core.Record {
	h := addRecord(t, a, "habits", map[string]any{"user": user, "title": "Vitamin D", "icon": "pill", "times": times})
	created := time.Now().AddDate(0, 0, -daysAgo).UTC().Format("2006-01-02 15:04:05.000Z")
	if _, err := a.DB().NewQuery("UPDATE habits SET created = {:c} WHERE id = {:id}").
		Bind(map[string]any{"c": created, "id": h.Id}).Execute(); err != nil {
		t.Fatal(err)
	}
	h, _ = a.FindRecordById("habits", h.Id)
	return h
}

func tick(t *testing.T, a core.App, h *core.Record, day string, count int) {
	t.Helper()
	c, err := a.FindFirstRecordByFilter("habit_checks", "habit = {:h} && day = {:d}", map[string]any{"h": h.Id, "d": day})
	if err != nil {
		col, _ := a.FindCollectionByNameOrId("habit_checks")
		c = core.NewRecord(col)
		c.Load(map[string]any{"habit": h.Id, "day": day})
	}
	c.Set("count", count)
	if err := a.Save(c); err != nil {
		t.Fatal(err)
	}
}

func berries(t *testing.T, a core.App, user string) int {
	t.Helper()
	n, err := rewards.Balance(a, user)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestBerries(t *testing.T) {
	today := habits.Today()

	t.Run("a habit done earns a berry, once, and a whole day three more", func(t *testing.T) {
		a := newApp(t)
		u, _ := addUser(t, a, "ada@example.com", "member")
		teeth := addHabit(t, a, u.Id, 2, 0)
		vitamin := addHabit(t, a, u.Id, 1, 0)
		tick(t, a, teeth, today, 1)
		if n := berries(t, a, u.Id); n != 0 {
			t.Fatalf("brushed once of twice: %d berries", n)
		}
		tick(t, a, teeth, today, 2)
		tick(t, a, teeth, today, 0)
		tick(t, a, teeth, today, 2)
		if n := berries(t, a, u.Id); n != 1 {
			t.Fatalf("ticked off twice: %d berries, want 1", n)
		}
		tick(t, a, vitamin, today, 1)
		if n := berries(t, a, u.Id); n != 1+1+rewards.PerDay {
			t.Fatalf("whole day done: %d berries", n)
		}
	})

	t.Run("ticking off last week doesn't pay", func(t *testing.T) {
		a := newApp(t)
		u, _ := addUser(t, a, "ada@example.com", "member")
		h := addHabit(t, a, u.Id, 1, 10)
		tick(t, a, h, habits.AddDays(today, -5), 1)
		if n := berries(t, a, u.Id); n != 0 {
			t.Fatalf("%d berries", n)
		}
	})

	t.Run("a week in a row earns a bonus and shows as a streak", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		h := addHabit(t, a, u.Id, 1, 10)
		for i := 6; i >= 0; i-- {
			tick(t, a, h, habits.AddDays(today, -i), 1)
		}
		// yesterday and today pay: 2 habits, 2 days, and the 7-day bonus
		want := 2*rewards.PerHabit + 2*rewards.PerDay + rewards.Milestones[7]
		(&tests.ApiScenario{
			Method:          http.MethodGet,
			URL:             "/api/app/play/me",
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"streak":7`, `"berries":` + itoa(want), `"lifetime":2`, `"stage":"cub"`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("a skipped day doesn't break the streak, a missed one does", func(t *testing.T) {
		a := newApp(t)
		u, _ := addUser(t, a, "ada@example.com", "member")
		h := addHabit(t, a, u.Id, 1, 10)
		for _, i := range []int{1, 3, 4} {
			tick(t, a, h, habits.AddDays(today, -i), 1)
		}
		skip := habits.AddDays(today, -2)
		tick(t, a, h, skip, 0)
		c, _ := a.FindFirstRecordByData("habit_checks", "day", skip)
		c.Set("skipped", true)
		if err := a.Save(c); err != nil {
			t.Fatal(err)
		}
		hist, err := habits.Load(a, u.Id, habits.AddDays(today, -30), today)
		if err != nil {
			t.Fatal(err)
		}
		// today isn't done yet; 1, (2 skipped), 3, 4 are; 5 was missed
		if n := hist.Streak(today, 30); n != 3 {
			t.Errorf("streak %d, want 3", n)
		}
	})
}

func TestShop(t *testing.T) {
	buy := func(t *testing.T, a *tests.TestApp, token, item string, status int, want string, after ...func(testing.TB, *tests.TestApp, *http.Response)) {
		s := &tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/app/play/buy",
			Body:            strings.NewReader(`{"item":"` + item + `"}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  status,
			ExpectedContent: []string{want},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}
		if len(after) > 0 {
			s.AfterTestFunc = after[0]
		}
		s.Test(t)
	}

	t.Run("not without enough berries", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		if err := rewards.Award(a, u.Id, 10, "habit", "test", ""); err != nil {
			t.Fatal(err)
		}
		buy(t, a, token, "bow_tie", http.StatusBadRequest, "Not enough berries")
	})

	t.Run("buying spends berries and puts it on", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		if err := rewards.Award(a, u.Id, 50, "habit", "test", ""); err != nil {
			t.Fatal(err)
		}
		addRecord(t, a, "inventory", map[string]any{"user": u.Id, "item": "beanie", "equipped": true})
		buy(t, a, token, "party_hat", http.StatusOK, `"item":"party_hat"`, func(t testing.TB, a *tests.TestApp, _ *http.Response) {
			if n, _ := rewards.Balance(a, u.Id); n != 30 {
				t.Errorf("%d berries left, want 30", n)
			}
			hat, _ := a.FindFirstRecordByData("inventory", "item", "party_hat")
			beanie, _ := a.FindFirstRecordByData("inventory", "item", "beanie")
			if !hat.GetBool("equipped") || beanie.GetBool("equipped") {
				t.Errorf("wearing the party hat %v and the beanie %v", hat.GetBool("equipped"), beanie.GetBool("equipped"))
			}
		})
	})

	t.Run("not twice", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		if err := rewards.Award(a, u.Id, 50, "habit", "test", ""); err != nil {
			t.Fatal(err)
		}
		addRecord(t, a, "inventory", map[string]any{"user": u.Id, "item": "bow_tie"})
		buy(t, a, token, "bow_tie", http.StatusBadRequest, "already")
	})

	t.Run("nobody writes the ledger but the server", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/ledger/records",
			Body:            strings.NewReader(`{"user":"` + u.Id + `","amount":1000,"reason":"habit","ref":"x"}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{`"status":403`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})
}

func itoa(n int) string { return strconv.Itoa(n) }
