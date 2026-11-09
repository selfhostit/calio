// Package rewards keeps the berries: earned for habits done, a whole day
// done and streaks, and spent in the shop. Every berry is a row in the
// ledger, written only here; a reason is earned once, by its unique ref.
package rewards

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/selfhostit/calio/play/habits"
)

// Ledger is the collection, from migration 1763400000.
const Ledger = "ledger"

// What earns berries.
const (
	PerHabit = 1
	PerDay   = 3
)

// Milestones are the streaks that earn a bonus, by length in days.
var Milestones = map[int]int{7: 10, 30: 30, 100: 100}

// Stages are how a companion grows with the habits done over time: a cub,
// then young, then grown up.
var Stages = []struct {
	Name string
	From int
}{{"cub", 0}, {"young", 30}, {"grown", 150}}

// ErrNotEnough is returned by Spend when someone can't afford it.
var ErrNotEnough = errors.New("not enough berries")

// how far back a streak is counted
const streakDays = 400

// Register earns berries as habits are ticked off and serves
// /api/app/play/me.
func Register(app core.App, h *habits.Module) {
	h.OnCheck.BindFunc(func(e *habits.CheckEvent) error {
		if err := earn(e); err != nil {
			return err
		}
		return e.Next()
	})
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/api/app/play/me", me).Bind(apis.RequireAuth("users"))
		return se.Next()
	})
}

func earn(e *habits.CheckEvent) error {
	// only for today (and yesterday, for late evenings and other time zones):
	// ticking off last month doesn't pay
	today := habits.Today()
	if e.Day < habits.AddDays(today, -1) || e.Day > habits.AddDays(today, 1) {
		return nil
	}
	if e.Done {
		if err := Award(e.App, e.User, PerHabit, "habit", "habit:"+e.Habit.Id+":"+e.Day, e.Day); err != nil {
			return err
		}
	}
	if !e.DayDone {
		return nil
	}
	if err := Award(e.App, e.User, PerDay, "day", "day:"+e.User+":"+e.Day, e.Day); err != nil {
		return err
	}
	h, err := habits.Load(e.App, e.User, habits.AddDays(e.Day, -streakDays), e.Day)
	if err != nil {
		return err
	}
	if bonus, ok := Milestones[h.Streak(e.Day, streakDays)]; ok {
		return Award(e.App, e.User, bonus, "streak", "streak:"+e.User+":"+e.Day, e.Day)
	}
	return nil
}

// Award gives someone berries for a reason, once per ref.
func Award(app core.App, user string, amount int, reason, ref, day string) error {
	if _, err := app.FindFirstRecordByData(Ledger, "ref", ref); err == nil {
		return nil
	}
	return add(app, user, amount, reason, ref, day)
}

// Spend takes berries for something, or returns ErrNotEnough. Run it in the
// same transaction as whatever is bought.
func Spend(tx core.App, user string, amount int, reason, ref string) error {
	balance, err := Balance(tx, user)
	if err != nil {
		return err
	}
	if balance < amount {
		return fmt.Errorf("%w: %d of %d", ErrNotEnough, balance, amount)
	}
	return add(tx, user, -amount, reason, ref, "")
}

func add(app core.App, user string, amount int, reason, ref, day string) error {
	c, err := app.FindCachedCollectionByNameOrId(Ledger)
	if err != nil {
		return err
	}
	r := core.NewRecord(c)
	r.Load(map[string]any{"user": user, "amount": amount, "reason": reason, "ref": ref, "day": day})
	return app.Save(r)
}

// Balance is how many berries someone has.
func Balance(app core.App, user string) (int, error) {
	return sum(app, dbx.HashExp{"user": user})
}

// Lifetime is how many habits someone has done, ever.
func Lifetime(app core.App, user string) (int, error) {
	var n int
	err := app.RecordQuery(Ledger).Select("count(*)").
		Where(dbx.HashExp{"user": user, "reason": "habit"}).Row(&n)
	return n, err
}

func sum(app core.App, where dbx.Expression) (int, error) {
	var n int
	err := app.RecordQuery(Ledger).Select("coalesce(sum(amount), 0)").Where(where).Row(&n)
	return n, err
}

// Stage is how grown a companion is after so many habits.
func Stage(lifetime int) string {
	stage := Stages[0].Name
	for _, s := range Stages {
		if lifetime >= s.From {
			stage = s.Name
		}
	}
	return stage
}

// me tells someone their berries, streak and how grown their companion is.
func me(e *core.RequestEvent) error {
	user := e.Auth.Id
	balance, err := Balance(e.App, user)
	if err != nil {
		return e.InternalServerError("", err)
	}
	lifetime, err := Lifetime(e.App, user)
	if err != nil {
		return e.InternalServerError("", err)
	}
	today := habits.Today()
	h, err := habits.Load(e.App, user, habits.AddDays(today, -streakDays), today)
	if err != nil {
		return e.InternalServerError("", err)
	}
	next := 0
	for _, s := range Stages {
		if s.From > lifetime {
			next = s.From
			break
		}
	}
	return e.JSON(http.StatusOK, map[string]any{
		"berries":  balance,
		"streak":   h.Streak(today, streakDays),
		"lifetime": lifetime,
		"stage":    Stage(lifetime),
		// habits done at which it grows next, or 0 when it's grown up
		"next_stage": next,
	})
}
