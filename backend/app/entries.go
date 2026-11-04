package app

import (
	"fmt"
	"strings"
	"time"

	validation "github.com/pocketbase/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"github.com/selfhostit/calio/jobs"
	"github.com/selfhostit/calio/push"
)

// How far ahead the morning digest warns about pet sitting nobody has booked.
const sitterWarning = 7 * 24 * time.Hour

var kindNames = map[string]string{"event": "an event", "petsit": "pet sitting", "trip": "a trip"}

func registerEntries(app core.App) {
	// Every save: the end can't come before the start, and pet sitting starts
	// out needing a sitter.
	app.OnRecordValidate("entries").BindFunc(func(e *core.RecordEvent) error {
		start, end := e.Record.GetDateTime("start"), e.Record.GetDateTime("end")
		if !end.IsZero() && end.Before(start) {
			return validation.Errors{"end": validation.NewError("end_before_start", "The end can't be before the start.")}
		}
		if e.Record.GetString("kind") == "petsit" && e.Record.GetString("sitter_status") == "" {
			e.Record.Set("sitter_status", "needed")
		}
		return e.Next()
	})

	// A new contact: remember who added it.
	app.OnRecordCreateRequest("contacts").BindFunc(func(e *core.RecordRequestEvent) error {
		if e.Auth != nil && e.Auth.Collection().Name == "users" {
			e.Record.Set("created_by", e.Auth.Id)
		}
		return e.Next()
	})

	// A new entry: remember who added it and tell everyone else.
	app.OnRecordCreateRequest("entries").BindFunc(func(e *core.RecordRequestEvent) error {
		byID, byName := "", "Someone"
		if e.Auth != nil && e.Auth.Collection().Name == "users" {
			byID, byName = e.Auth.Id, displayName(e.Auth)
			e.Record.Set("created_by", byID)
		}
		if err := e.Next(); err != nil {
			return err
		}
		push.ToEveryone(e.App, byID, push.Message{
			Title: fmt.Sprintf("%s added %s", byName, kindNames[e.Record.GetString("kind")]),
			Body:  e.Record.GetString("title") + ", " + when(e.Record),
			Tag:   "entry-" + e.Record.Id,
		})
		return nil
	})

	// A sitter is booked: good news for everyone.
	app.OnRecordUpdateRequest("entries").BindFunc(func(e *core.RecordRequestEvent) error {
		wasBooked := e.Record.Original().GetString("sitter_status") == "booked"
		if err := e.Next(); err != nil {
			return err
		}
		if !wasBooked && e.Record.GetString("sitter_status") == "booked" {
			byID := ""
			if e.Auth != nil {
				byID = e.Auth.Id
			}
			who := e.Record.GetString("sitter")
			if who == "" {
				who = "A sitter"
			}
			push.ToEveryone(e.App, byID, push.Message{
				Title: "Pet sitter booked",
				Body:  fmt.Sprintf("%s is covering %s", who, when(e.Record)),
				Tag:   "entry-" + e.Record.Id,
			})
		}
		return nil
	})

	// Every morning: what's on today, and pet sitting that still needs someone.
	jobs.MustAdd(app, "morning-digest", "30 7 * * *", func(app core.App) error {
		msg, ok, err := morningDigest(app, time.Now())
		if err != nil || !ok {
			return err
		}
		push.ToEveryone(app, "", msg)
		return nil
	})
}

// morningDigest sums up the day; ok is false when there is nothing to say.
func morningDigest(app core.App, now time.Time) (push.Message, bool, error) {
	today := now.In(time.Local).Format(time.DateOnly)
	// Wide bounds in UTC; the exact day check happens per entry below.
	from := now.Add(-400 * 24 * time.Hour).UTC().Format(types.DefaultDateLayout)
	to := now.Add(sitterWarning + 48*time.Hour).UTC().Format(types.DefaultDateLayout)
	entries, err := app.FindRecordsByFilter("entries",
		"start <= {:to} && (end >= {:from} || start >= {:from} || repeat != '')", "start", 0, 0,
		map[string]any{"from": from, "to": to})
	if err != nil {
		return push.Message{}, false, err
	}

	var on, sitters []string
	for _, r := range entries {
		first, _ := days(r)
		if occursOn(r, today) {
			on = append(on, r.GetString("title"))
		}
		if r.GetString("kind") == "petsit" && r.GetString("repeat") == "" &&
			r.GetString("sitter_status") != "booked" && first >= today {
			if start := r.GetDateTime("start").Time(); start.Sub(now) <= sitterWarning {
				sitters = append(sitters, r.GetString("title")+" ("+when(r)+")")
			}
		}
	}
	if len(on) == 0 && len(sitters) == 0 {
		return push.Message{}, false, nil
	}

	var body []string
	if len(on) > 0 {
		body = append(body, "Today: "+strings.Join(on, ", "))
	}
	if len(sitters) > 0 {
		body = append(body, "Still needs a pet sitter: "+strings.Join(sitters, ", "))
	}
	title := "Good morning"
	if len(on) > 0 {
		title = fmt.Sprintf("%d on today", len(on))
	}
	return push.Message{Title: title, Body: strings.Join(body, ". "), Tag: "morning"}, true, nil
}

// days gives an entry's first and last calendar day (YYYY-MM-DD) in the
// server's time zone. All-day dates are kept as midnight UTC, so they are read
// as they are.
func days(r *core.Record) (string, string) {
	start, end := r.GetDateTime("start").Time(), r.GetDateTime("end").Time()
	if end.IsZero() {
		end = start
	}
	if r.GetBool("all_day") {
		return start.UTC().Format(time.DateOnly), end.UTC().Format(time.DateOnly)
	}
	return start.In(time.Local).Format(time.DateOnly), end.In(time.Local).Format(time.DateOnly)
}

// occursOn reports whether an entry, or one of its repeats, is on a day
// (YYYY-MM-DD in the server's time zone).
func occursOn(r *core.Record, day string) bool {
	first, last := days(r)
	if first <= day && day <= last {
		return true
	}
	rule := r.GetString("repeat")
	if rule == "" || day < first {
		return false
	}
	until := "9999-12-31"
	if u := r.GetDateTime("repeat_until"); !u.IsZero() {
		until = u.Time().UTC().Format(time.DateOnly)
	}
	start, _ := time.Parse(time.DateOnly, first)
	end, _ := time.Parse(time.DateOnly, last)
	length := end.Sub(start)
	for n := 1; n < 100_000; n++ {
		occ, ok := nthRepeat(start, rule, n)
		if !ok {
			continue
		}
		d := occ.Format(time.DateOnly)
		if d > day || d > until {
			return false
		}
		if day <= occ.Add(length).Format(time.DateOnly) {
			return true
		}
	}
	return false
}

// nthRepeat gives the day of the nth repeat after start. ok is false for a
// month or year that doesn't have that day (the 31st, or 29 February), which
// is skipped, as calendar apps do.
func nthRepeat(start time.Time, rule string, n int) (time.Time, bool) {
	var t time.Time
	switch rule {
	case "daily":
		return start.AddDate(0, 0, n), true
	case "weekly":
		return start.AddDate(0, 0, 7*n), true
	case "monthly":
		t = start.AddDate(0, n, 0)
	case "yearly":
		t = start.AddDate(n, 0, 0)
	default:
		return t, false
	}
	return t, t.Day() == start.Day()
}

// when describes an entry's dates for a notification, like "Fri 12 Dec" or
// "Fri 12 – Sun 14 Dec" or "Fri 12 Dec 18:30".
func when(r *core.Record) string {
	start, end := r.GetDateTime("start").Time(), r.GetDateTime("end").Time()
	if r.GetBool("all_day") {
		start, end = start.UTC(), end.UTC()
	} else {
		start, end = start.In(time.Local), end.In(time.Local)
	}
	first, last := days(r)
	switch {
	case first == last && r.GetBool("all_day"):
		return start.Format("Mon 2 Jan")
	case first == last:
		return start.Format("Mon 2 Jan 15:04")
	case start.Month() == end.Month() && start.Year() == end.Year():
		return start.Format("Mon 2") + " – " + end.Format("Mon 2 Jan")
	default:
		return start.Format("Mon 2 Jan") + " – " + end.Format("Mon 2 Jan")
	}
}

func displayName(u *core.Record) string {
	if n := strings.TrimSpace(u.GetString("name")); n != "" {
		return strings.Fields(n)[0]
	}
	return strings.Split(u.Email(), "@")[0]
}
