package app

import (
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

// The calendar feed: an iCalendar file of everything on the calendar that
// phone and desktop calendar apps subscribe to. Calendar apps can't sign in,
// so each person's feed URL carries a secret (users.feed_token).
//
// Only GET /api/app/feed/{token}.ics lives under /api/app/feed/: an optional
// Tailscale Funnel publishes that prefix, and nothing else, to the internet so
// Google Calendar can reach it. CALIO_FEED_URL is then that public address.
const feedPrefix = "/api/app/feed/"

func registerFeedRoutes(se *core.ServeEvent) {
	se.Router.GET(feedPrefix+"{file}", serveFeed)

	link := se.Router.Group("/api/app/feed-link").Bind(apis.RequireAuth("users"))
	link.GET("", func(e *core.RequestEvent) error { return feedLink(e, false) })
	link.POST("/reset", func(e *core.RequestEvent) error { return feedLink(e, true) })
}

// feedLink tells a person their feed URL, making a secret on first use or
// when they ask for a new one (which stops the old URL working).
func feedLink(e *core.RequestEvent, reset bool) error {
	token := e.Auth.GetString("feed_token")
	if token == "" || reset {
		token = security.RandomString(32)
		e.Auth.Set("feed_token", token)
		if err := e.App.Save(e.Auth); err != nil {
			return e.InternalServerError("", err)
		}
	}
	base, public := strings.TrimRight(os.Getenv("CALIO_FEED_URL"), "/"), true
	if base == "" {
		base, public = e.Request.Header.Get("Origin"), false
		if base == "" {
			scheme := "http"
			if e.IsTLS() || e.Request.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			base = scheme + "://" + e.Request.Host
		}
	}
	return e.JSON(http.StatusOK, map[string]any{
		"url":    base + feedPrefix + token + ".ics",
		"public": public,
	})
}

func serveFeed(e *core.RequestEvent) error {
	token, ok := strings.CutSuffix(e.Request.PathValue("file"), ".ics")
	if !ok || len(token) < 32 {
		return e.NotFoundError("", nil)
	}
	if _, err := e.App.FindFirstRecordByData("users", "feed_token", token); err != nil {
		return e.NotFoundError("", nil)
	}
	entries, err := e.App.FindRecordsByFilter("entries", "", "start", 0, 0)
	if err != nil {
		return e.InternalServerError("", err)
	}
	h := e.Response.Header()
	h.Set("Content-Type", "text/calendar; charset=utf-8")
	h.Set("Cache-Control", "private, max-age=300")
	return e.String(http.StatusOK, icalendar(entries, time.Now()))
}

var kindPrefix = map[string]string{"petsit": "🐾 ", "trip": "✈️ "}

// icalendar writes entries as an iCalendar (RFC 5545) document.
func icalendar(entries []*core.Record, now time.Time) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(fold(s)) }
	line("BEGIN:VCALENDAR")
	line("VERSION:2.0")
	line("PRODID:-//selfhostit//Calio//EN")
	line("CALSCALE:GREGORIAN")
	line("METHOD:PUBLISH")
	line("X-WR-CALNAME:Calio")
	line("REFRESH-INTERVAL;VALUE=DURATION:PT1H")
	line("X-PUBLISHED-TTL:PT1H")
	for _, r := range entries {
		line("BEGIN:VEVENT")
		line("UID:" + r.Id + "@calio")
		stamp := r.GetDateTime("updated").Time()
		if stamp.IsZero() {
			stamp = now
		}
		line("DTSTAMP:" + stamp.UTC().Format("20060102T150405Z"))

		summary := kindPrefix[r.GetString("kind")] + r.GetString("title")
		if r.GetString("kind") == "petsit" {
			summary += " (" + sitterLine(r) + ")"
		}
		line("SUMMARY:" + escape(summary))

		start, end := r.GetDateTime("start").Time(), r.GetDateTime("end").Time()
		allDay := r.GetBool("all_day")
		if allDay {
			first, last := days(r)
			s, _ := time.Parse(time.DateOnly, first)
			l, _ := time.Parse(time.DateOnly, last)
			line("DTSTART;VALUE=DATE:" + s.Format("20060102"))
			line("DTEND;VALUE=DATE:" + l.AddDate(0, 0, 1).Format("20060102")) // the day after
		} else {
			if end.IsZero() || end.Before(start) {
				end = start.Add(time.Hour)
			}
			line("DTSTART:" + start.UTC().Format("20060102T150405Z"))
			line("DTEND:" + end.UTC().Format("20060102T150405Z"))
		}

		if rule := r.GetString("repeat"); rule != "" {
			rrule := "RRULE:FREQ=" + strings.ToUpper(rule)
			if until := r.GetDateTime("repeat_until").Time(); !until.IsZero() {
				if allDay {
					rrule += ";UNTIL=" + until.UTC().Format("20060102")
				} else {
					// the end of that day where the server is
					d := until.UTC()
					endOfDay := time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, time.Local)
					rrule += ";UNTIL=" + endOfDay.UTC().Format("20060102T150405Z")
				}
			}
			line(rrule)
		}
		if loc := r.GetString("location"); loc != "" {
			line("LOCATION:" + escape(loc))
		}
		if desc := description(r); desc != "" {
			line("DESCRIPTION:" + escape(desc))
		}
		line("END:VEVENT")
	}
	line("END:VCALENDAR")
	return b.String()
}

func sitterLine(r *core.Record) string {
	sitter, status := r.GetString("sitter"), r.GetString("sitter_status")
	switch {
	case status == "booked" && sitter != "":
		return sitter + " is booked"
	case status == "booked":
		return "sitter booked"
	case status == "asked" && sitter != "":
		return "asked " + sitter
	default:
		return "sitter needed"
	}
}

func description(r *core.Record) string {
	var parts []string
	if n := strings.TrimSpace(r.GetString("notes")); n != "" {
		parts = append(parts, n)
	}
	var items []struct {
		Text string `json:"text"`
		Done bool   `json:"done"`
	}
	if err := r.UnmarshalJSONField("checklist", &items); err == nil && len(items) > 0 {
		var list []string
		for _, i := range items {
			mark := "☐"
			if i.Done {
				mark = "☑"
			}
			list = append(list, mark+" "+i.Text)
		}
		parts = append(parts, strings.Join(list, "\n"))
	}
	return strings.Join(parts, "\n\n")
}

// escape makes text safe for an iCalendar value.
func escape(s string) string {
	return strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\r\n", `\n`, "\n", `\n`).Replace(s)
}

// fold ends a content line, breaking it so no line is longer than 75 bytes
// and never inside a UTF-8 character.
func fold(s string) string {
	var b strings.Builder
	limit := 75
	for len(s) > limit {
		cut := limit
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		b.WriteString(s[:cut] + "\r\n ")
		s = s[cut:]
		limit = 74 // the leading space counts
	}
	b.WriteString(s + "\r\n")
	return b.String()
}
