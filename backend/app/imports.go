package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"github.com/selfhostit/calio/ical"
	"github.com/selfhostit/calio/jobs"
)

// Calendar import: a .ics file someone uploads (POST /api/app/import), and
// calendars the household subscribes to by their link (the calendars
// collection), which the server refreshes every half hour.
//
// Imported events are plain entries, so they can be commented on, given a
// sitter or a packing list. Importing again, or a refresh, updates the title,
// times, place and notes from the source and keeps everything else.

const (
	// events that ended longer ago than this are left out
	importPast = 365 * 24 * time.Hour
	// a repeat Calio can't show as one entry becomes one entry per time, this far ahead
	importAhead = 2 * 365 * 24 * time.Hour
	// the most entries one repeating event becomes
	importMaxRepeats = 500
	// the largest calendar file read
	importMaxBytes = 20 << 20
)

func registerImports(app core.App) {
	jobs.MustAdd(app, "calendar-refresh", "*/30 * * * *", func(app core.App) error {
		calendars, err := app.FindAllRecords("calendars")
		if err != nil {
			return err
		}
		var failed []string
		for _, c := range calendars {
			if err := syncCalendar(app, c, time.Now(), false); err != nil {
				failed = append(failed, c.GetString("name")+": "+err.Error())
			}
		}
		if len(failed) > 0 {
			return errors.New(strings.Join(failed, "; "))
		}
		return nil
	})
}

func registerImportRoutes(se *core.ServeEvent) {
	se.Router.POST("/api/app/import", importFile).Bind(apis.RequireAuth("users"), apis.BodyLimit(importMaxBytes))
	se.Router.POST("/api/app/calendars", subscribe).Bind(apis.RequireAuth("users"))
	se.Router.POST("/api/app/calendars/{id}/refresh", func(e *core.RequestEvent) error {
		c, err := e.App.FindRecordById("calendars", e.Request.PathValue("id"))
		if err != nil {
			return e.NotFoundError("That calendar is gone.", err)
		}
		syncCalendar(e.App, c, time.Now(), false)
		return e.JSON(http.StatusOK, c)
	}).Bind(apis.RequireAuth("users"))
}

// subscribe adds a calendar by its link and fetches it right away, so its
// events show up. The link is hidden, so only this route sets it; to change
// it, remove the calendar and add it again.
func subscribe(e *core.RequestEvent) error {
	var body struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Category string `json:"category"`
	}
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("Paste the calendar's link.", err)
	}
	link, err := normalizeLink(body.URL)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if body.Category != "" {
		if _, err := e.App.FindRecordById("categories", body.Category); err != nil {
			return e.BadRequestError("That category is gone.", err)
		}
	}
	col, err := e.App.FindCollectionByNameOrId("calendars")
	if err != nil {
		return e.InternalServerError("", err)
	}
	c := core.NewRecord(col)
	c.Set("url", link)
	c.Set("category", body.Category)
	c.Set("created_by", e.Auth.Id)
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "Calendar"
	}
	c.Set("name", truncate(name, 80))
	if err := e.App.Save(c); err != nil {
		return e.BadRequestError("Couldn't add that calendar.", err)
	}
	// without a name of its own, it takes the one the calendar gives itself
	syncCalendar(e.App, c, time.Now(), strings.TrimSpace(body.Name) == "")
	return e.JSON(http.StatusOK, c)
}

// normalizeLink turns webcal:// links into https:// ones and checks the link
// is a web address.
func normalizeLink(link string) (string, error) {
	link = strings.TrimSpace(link)
	lower := strings.ToLower(link)
	for _, scheme := range []string{"webcals://", "webcal://"} {
		if strings.HasPrefix(lower, scheme) {
			link = "https://" + link[len(scheme):]
			break
		}
	}
	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("That doesn't look like a calendar link. It starts with https:// or webcal://.")
	}
	return link, nil
}

// importFile adds the events in an uploaded .ics file to the calendar.
func importFile(e *core.RequestEvent) error {
	file, _, err := e.Request.FormFile("file")
	if err != nil {
		return e.BadRequestError("Choose a calendar file (.ics) to import.", err)
	}
	defer file.Close()
	cal, err := ical.Parse(io.LimitReader(file, importMaxBytes), time.Local)
	if err != nil {
		return e.BadRequestError("That file isn't a calendar (.ics) file.", err)
	}
	category := e.Request.FormValue("category")
	if category != "" {
		if _, err := e.App.FindRecordById("categories", category); err != nil {
			return e.BadRequestError("That category is gone.", err)
		}
	}
	res, err := saveEvents(e.App, cal, importTarget{category: category, by: e.Auth.Id}, time.Now())
	if err != nil {
		return e.InternalServerError("", err)
	}
	return e.JSON(http.StatusOK, res)
}

type importTarget struct {
	calendar string // the subscription, or "" for a file
	category string
	by       string
}

type importResult struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Removed int `json:"removed"`
	// Total is how many entries the calendar has in Calio now.
	Total int `json:"total"`
}

// syncCalendar fetches a subscription and brings its entries up to date. The
// outcome is saved on the calendar (synced, error, events), and an error is
// returned for the job's record.
func syncCalendar(app core.App, c *core.Record, now time.Time, rename bool) error {
	res, err := fetchAndSave(app, c, now, rename)
	if err != nil {
		c.Set("error", truncate(err.Error(), 500))
	} else {
		c.Set("error", "")
		c.Set("events", res.Total)
	}
	c.Set("synced", now)
	if saveErr := app.Save(c); saveErr != nil && err == nil {
		err = saveErr
	}
	return err
}

func fetchAndSave(app core.App, c *core.Record, now time.Time, rename bool) (importResult, error) {
	cal, err := fetchCalendar(c.GetString("url"))
	if err != nil {
		return importResult{}, err
	}
	if name := strings.TrimSpace(cal.Name); rename && name != "" {
		c.Set("name", truncate(name, 80))
	}
	return saveEvents(app, cal, importTarget{
		calendar: c.Id, category: c.GetString("category"), by: c.GetString("created_by"),
	}, now)
}

var importClient = &http.Client{Timeout: 60 * time.Second}

func fetchCalendar(link string) (*ical.Calendar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, errors.New("that link isn't a web address")
	}
	req.Header.Set("Accept", "text/calendar, */*")
	req.Header.Set("User-Agent", "Calio/"+Version)
	res, err := importClient.Do(req)
	if err != nil {
		// without the link, which is a secret
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errors.New("the calendar took too long to answer")
		}
		return nil, errors.New("couldn't reach the calendar (" + err.Error() + ")")
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusGone:
		return nil, errors.New("the calendar isn't there any more (was the link reset or the calendar unpublished?)")
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return nil, errors.New("the calendar wouldn't let Calio in; use its private or public iCal link")
	case res.StatusCode >= 300:
		return nil, fmt.Errorf("the calendar answered %s", res.Status)
	}
	cal, err := ical.Parse(io.LimitReader(res.Body, importMaxBytes), time.Local)
	if err != nil {
		return nil, errors.New("the link doesn't lead to a calendar (.ics); copy the iCal or ICS link")
	}
	return cal, nil
}

// item is one entry an imported event becomes.
type item struct {
	uid, title, notes, location string
	allDay                      bool
	start, end                  time.Time
	repeat                      string
	repeatUntil                 time.Time
}

// saveEvents adds and updates the entries for a calendar's events. For a
// subscription it also removes the entries of events that are gone from it.
func saveEvents(app core.App, cal *ical.Calendar, to importTarget, now time.Time) (importResult, error) {
	items := toItems(cal.Events, now)
	var res importResult
	err := app.RunInTransaction(func(tx core.App) error {
		category := to.category
		if category == "" {
			category = defaultCategory(tx)
		}
		// birthdays go in the birthday category, unless the calendar was put
		// in a trip or sitting one
		birthdays := ""
		if plainCategory(tx, category) {
			birthdays = BirthdayCategory(tx)
		}
		wholeCalendar := birthdays != "" && IsBirthday(cal.Name)
		existing, err := tx.FindRecordsByFilter("entries", "uid != '' && calendar = {:c}", "", 0, 0,
			map[string]any{"c": to.calendar})
		if err != nil {
			return err
		}
		byUID := map[string]*core.Record{}
		for _, r := range existing {
			byUID[r.GetString("uid")] = r
		}
		entries, err := tx.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, it := range items {
			if seen[it.uid] {
				continue
			}
			seen[it.uid] = true
			r, ok := byUID[it.uid]
			if !ok {
				r = core.NewRecord(entries)
				r.Set("uid", it.uid)
				r.Set("calendar", to.calendar)
				r.Set("kind", "event")
				r.Set("category", category)
				if birthdays != "" && (wholeCalendar || IsBirthday(it.title)) {
					r.Set("category", birthdays)
				}
				r.Set("created_by", to.by)
			}
			if !fill(r, it) && ok {
				continue
			}
			if err := tx.Save(r); err != nil {
				return fmt.Errorf("couldn't save %q: %w", it.title, err)
			}
			if ok {
				res.Updated++
			} else {
				res.Added++
			}
		}
		if to.calendar != "" {
			for uid, r := range byUID {
				if !seen[uid] {
					if err := tx.Delete(r); err != nil {
						return err
					}
					res.Removed++
				}
			}
			res.Total = len(seen)
		} else {
			res.Total = res.Added + res.Updated
		}
		return nil
	})
	return res, err
}

// fill sets an entry's fields from an event, telling whether anything changed.
func fill(r *core.Record, it item) bool {
	start := dbTime(it.start, it.allDay)
	end := ""
	if it.allDay {
		if last := it.end.AddDate(0, 0, -1); last.After(it.start) {
			end = dbTime(last, true)
		}
	} else if it.end.After(it.start) {
		end = dbTime(it.end, false)
	}
	until := ""
	if !it.repeatUntil.IsZero() {
		until = dbTime(it.repeatUntil, true)
	}
	values := map[string]any{
		"title": it.title, "notes": it.notes, "location": it.location, "all_day": it.allDay,
		"start": start, "end": end, "repeat": it.repeat, "repeat_until": until,
	}
	changed := false
	for k, v := range values {
		var now any = r.GetString(k)
		switch k {
		case "all_day":
			now = r.GetBool(k)
		case "start", "end", "repeat_until":
			now = ""
			if t := r.GetDateTime(k); !t.IsZero() {
				now = t.String()
			}
		}
		if now != v {
			r.Set(k, v)
			changed = true
		}
	}
	return changed
}

// dbTime writes a time the way PocketBase stores it; all-day dates at
// midnight UTC.
func dbTime(t time.Time, allDay bool) string {
	if allDay {
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
	dt, _ := types.ParseDateTime(t.UTC())
	return dt.String()
}

// defaultCategory is where imported events go when no category was picked:
// the first plain category that is switched on.
func defaultCategory(app core.App) string {
	categories, err := app.FindRecordsByFilter("categories", "kind = 'event'", "-enabled,sort,created", 1, 0)
	if err != nil || len(categories) == 0 {
		return ""
	}
	return categories[0].Id
}

// toItems turns events into entries. A repeat Calio can show (every day,
// week, month or year, until a day) stays one entry; a weekly repeat on
// several days becomes one entry per day; any other repeat, or one with
// skipped or changed times, becomes one entry per time over the next two
// years. Cancelled events and events that ended over a year ago are left out.
func toItems(events []ical.Event, now time.Time) []item {
	from, to := now.Add(-importPast), now.Add(importAhead)
	type group struct {
		master    *ical.Event
		overrides []ical.Event
	}
	groups := map[string]*group{}
	var order []string
	for i := range events {
		ev := events[i]
		uid := ev.UID
		if uid == "" {
			uid = fmt.Sprintf("%s|%s", ev.Summary, ev.Start.UTC().Format("20060102T150405Z"))
		}
		g, ok := groups[uid]
		if !ok {
			g = &group{}
			groups[uid] = g
			order = append(order, uid)
		}
		if ev.RecurrenceID.IsZero() {
			ev.UID = uid
			g.master = &ev
		} else {
			ev.UID = uid
			g.overrides = append(g.overrides, ev)
		}
	}

	var out []item
	for _, uid := range order {
		g := groups[uid]
		m := g.master
		if m == nil {
			// changed times of an event that isn't in the file
			for _, o := range g.overrides {
				if !o.Cancelled && o.End.After(from) {
					out = append(out, itemOf(o, occurrenceUID(uid, o.RecurrenceID)))
				}
			}
			continue
		}
		if m.Cancelled {
			continue
		}
		if m.Rule == nil {
			if m.End.After(from) {
				out = append(out, itemOf(*m, uid))
			}
			continue
		}
		if len(g.overrides) == 0 && len(m.ExDates) == 0 {
			if repeat, ok := simpleRepeat(*m); ok {
				until, ended := repeatEnd(*m)
				if ended && until.Before(from) {
					continue
				}
				it := itemOf(*m, uid)
				it.repeat, it.repeatUntil = repeat, until
				out = append(out, it)
				continue
			}
			if days, ok := weeklyOnDays(*m); ok {
				for _, wd := range days {
					ev := *m
					shift := (int(wd) - int(m.Start.Weekday()) + 7) % 7
					ev.Start, ev.End = m.Start.AddDate(0, 0, shift), m.End.AddDate(0, 0, shift)
					ev.Rule = &ical.Rule{Freq: "WEEKLY", Interval: 1, Until: m.Rule.Until}
					until, ended := repeatEnd(ev)
					if ended && until.Before(from) {
						continue
					}
					it := itemOf(ev, uid+"#"+strings.ToUpper(wd.String()[:2]))
					it.repeat, it.repeatUntil = "weekly", until
					out = append(out, it)
				}
				continue
			}
		}
		// one entry per time
		moved := map[int64]ical.Event{}
		for _, o := range g.overrides {
			moved[o.RecurrenceID.Unix()] = o
		}
		for _, t := range m.Occurrences(from, to, importMaxRepeats) {
			ev := *m
			if o, ok := moved[t.Unix()]; ok {
				delete(moved, t.Unix())
				if o.Cancelled {
					continue
				}
				ev = o
			} else {
				ev.End = t.Add(m.End.Sub(m.Start))
				ev.Start = t
			}
			ev.Rule = nil
			out = append(out, itemOf(ev, occurrenceUID(uid, t)))
		}
		for _, o := range moved {
			if !o.Cancelled && o.End.After(from) && o.Start.Before(to) {
				out = append(out, itemOf(o, occurrenceUID(uid, o.RecurrenceID)))
			}
		}
	}
	return out
}

func occurrenceUID(uid string, t time.Time) string {
	return uid + "#" + t.UTC().Format("20060102T150405Z")
}

func itemOf(ev ical.Event, uid string) item {
	title := ev.Summary
	if title == "" {
		title = "Busy"
	}
	return item{
		uid:      truncate(uid, 500),
		title:    truncate(title, 200),
		notes:    truncate(ev.Description, 2000),
		location: truncate(ev.Location, 200),
		allDay:   ev.AllDay,
		start:    ev.Start,
		end:      ev.End,
	}
}

// simpleRepeat tells whether an event repeats in a way one Calio entry shows:
// every day, week, month or year on the day it starts.
func simpleRepeat(ev ical.Event) (string, bool) {
	r := ev.Rule
	if r.Other || r.Interval != 1 || len(r.ByMonthDay) > 1 || len(r.ByMonth) > 1 {
		return "", false
	}
	sameDay := len(r.ByMonthDay) == 0 || r.ByMonthDay[0] == ev.Start.Day()
	sameMonth := len(r.ByMonth) == 0 || r.ByMonth[0] == int(ev.Start.Month())
	switch r.Freq {
	case "DAILY":
		return "daily", len(r.ByDay) == 0 && len(r.ByMonthDay) == 0 && len(r.ByMonth) == 0
	case "WEEKLY":
		onStart := len(r.ByDay) == 0 || (len(r.ByDay) == 1 && r.ByDay[0].N == 0 && r.ByDay[0].Day == ev.Start.Weekday())
		return "weekly", onStart && len(r.ByMonthDay) == 0 && len(r.ByMonth) == 0
	case "MONTHLY":
		return "monthly", len(r.ByDay) == 0 && sameDay && len(r.ByMonth) == 0
	case "YEARLY":
		return "yearly", len(r.ByDay) == 0 && sameDay && sameMonth
	}
	return "", false
}

// weeklyOnDays tells the days of a weekly repeat on several days ("every
// Monday and Wednesday").
func weeklyOnDays(ev ical.Event) ([]time.Weekday, bool) {
	r := ev.Rule
	if r.Freq != "WEEKLY" || r.Other || r.Interval != 1 || r.Count > 0 ||
		len(r.ByDay) < 2 || len(r.ByMonthDay) > 0 || len(r.ByMonth) > 0 {
		return nil, false
	}
	var days []time.Weekday
	for _, d := range r.ByDay {
		if d.N != 0 {
			return nil, false
		}
		days = append(days, d.Day)
	}
	return days, true
}

// repeatEnd is the last day a repeat happens on (as a date at midnight UTC),
// and whether it ends at all.
func repeatEnd(ev ical.Event) (time.Time, bool) {
	r := ev.Rule
	var last time.Time
	switch {
	case r.Count > 0:
		times := ev.Occurrences(ev.Start, ev.Start.AddDate(1000, 0, 0), r.Count)
		if len(times) == 0 {
			return time.Time{}, false
		}
		last = times[len(times)-1]
	case !r.Until.IsZero():
		last = r.Until
	default:
		return time.Time{}, false
	}
	if !ev.AllDay {
		last = last.In(time.Local)
	}
	return time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC), true
}

// truncate shortens s to at most n characters.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
