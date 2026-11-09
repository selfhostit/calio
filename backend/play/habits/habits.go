// Package habits knows when a habit is due and whether a day's habits are
// done, and tells the rest of play when a habit is ticked off.
//
// The web app has the same rules (web/src/lib/play/habits.ts); keep them in
// step.
package habits

import (
	"slices"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

// The collections, from migration 1763300000.
const (
	Habits = "habits"
	Checks = "habit_checks"
)

// DayFormat is how a day is written: a plain date in the household's time.
const DayFormat = "2006-01-02"

// CheckEvent is a habit's check for a day, just saved.
type CheckEvent struct {
	hook.Event
	App   core.App
	User  string
	Habit *core.Record
	Check *core.Record
	Day   string
	// Done is whether the habit is done that day, as many times as it's due.
	Done bool
	// DayDone is whether everything due that day is done.
	DayDone bool
}

// Module is the habits part of play.
type Module struct {
	// OnCheck runs after a check is saved; a handler's error is logged, never
	// shown to the person ticking.
	OnCheck hook.Hook[*CheckEvent]
}

// Register follows saved checks and returns the module, for others to
// listen to.
func Register(app core.App) *Module {
	m := &Module{}
	after := func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if err := m.checked(e.App, e.Record); err != nil {
			e.App.Logger().Error("After a habit check", "check", e.Record.Id, "error", err)
		}
		return nil
	}
	app.OnRecordAfterCreateSuccess(Checks).BindFunc(after)
	app.OnRecordAfterUpdateSuccess(Checks).BindFunc(after)
	return m
}

func (m *Module) checked(app core.App, check *core.Record) error {
	habit, err := app.FindRecordById(Habits, check.GetString("habit"))
	if err != nil {
		return err
	}
	user, day := habit.GetString("user"), check.GetString("day")
	h, err := Load(app, user, day, day)
	if err != nil {
		return err
	}
	return m.OnCheck.Trigger(&CheckEvent{
		App:     app,
		User:    user,
		Habit:   habit,
		Check:   check,
		Day:     day,
		Done:    !check.GetBool("skipped") && check.GetInt("count") >= habit.GetInt("times"),
		DayDone: h.DayDone(day),
	})
}

// History is someone's habits and their checks over some days.
type History struct {
	habits []*core.Record
	// checks by habit id and day
	checks map[string]*core.Record
}

// Load reads someone's habits and their checks from one day to another.
func Load(app core.App, user, from, to string) (*History, error) {
	habits, err := app.FindAllRecords(Habits, dbx.HashExp{"user": user})
	if err != nil {
		return nil, err
	}
	var checks []*core.Record
	if len(habits) > 0 {
		ids := make([]any, len(habits))
		for i, h := range habits {
			ids[i] = h.Id
		}
		checks, err = app.FindAllRecords(Checks,
			dbx.In("habit", ids...),
			dbx.Between("day", from, to),
		)
		if err != nil {
			return nil, err
		}
	}
	h := &History{habits: habits, checks: map[string]*core.Record{}}
	for _, c := range checks {
		h.checks[c.GetString("habit")+"/"+c.GetString("day")] = c
	}
	return h, nil
}

func (h *History) check(habit *core.Record, day string) *core.Record {
	return h.checks[habit.Id+"/"+day]
}

// Due reports whether a habit is something to do on a day: in use, added by
// then, on one of its weekdays and not skipped.
func (h *History) Due(habit *core.Record, day string) bool {
	if habit.GetBool("archived") || day < Since(habit) {
		return false
	}
	if days := habit.GetStringSlice("days"); len(days) > 0 && !slices.Contains(days, weekday(day)) {
		return false
	}
	c := h.check(habit, day)
	return c == nil || !c.GetBool("skipped")
}

// DayDone reports whether something was due on a day and all of it was done.
func (h *History) DayDone(day string) bool {
	due := false
	for _, habit := range h.habits {
		if !h.Due(habit, day) {
			continue
		}
		due = true
		c := h.check(habit, day)
		if c == nil || c.GetInt("count") < habit.GetInt("times") {
			return false
		}
	}
	return due
}

// Streak is how many days in a row up to a day had everything done. Days
// with nothing due don't break it, and the day itself only counts once it's
// done, so this morning doesn't end yesterday's streak.
func (h *History) Streak(day string, back int) int {
	n := 0
	for i := 0; i <= back; i++ {
		d := AddDays(day, -i)
		if h.DayDone(d) {
			n++
			continue
		}
		if i == 0 || !h.anyDue(d) {
			continue
		}
		break
	}
	return n
}

func (h *History) anyDue(day string) bool {
	for _, habit := range h.habits {
		if h.Due(habit, day) {
			return true
		}
	}
	return false
}

// Since is the day a habit was added, from which on it counts.
func Since(habit *core.Record) string {
	return habit.GetDateTime("created").Time().In(time.Local).Format(DayFormat)
}

// Today is the day it is now, in the household's time.
func Today() string { return time.Now().Format(DayFormat) }

// AddDays moves a day by n days.
func AddDays(day string, n int) string {
	t, err := time.Parse(DayFormat, day)
	if err != nil {
		return day
	}
	return t.AddDate(0, 0, n).Format(DayFormat)
}

var weekdays = []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

func weekday(day string) string {
	t, err := time.Parse(DayFormat, day)
	if err != nil {
		return ""
	}
	return weekdays[t.Weekday()]
}
