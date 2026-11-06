// Package ical reads the events in iCalendar (RFC 5545) files: the .ics files
// that Google Calendar, Outlook, Apple Calendar and others export, and the
// calendar links they publish. It reads what a household calendar needs (the
// title, the times, the place, the notes and how an event repeats) and skips
// the rest.
package ical

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Calendar is what an iCalendar file holds.
type Calendar struct {
	// Name is the calendar's own name (X-WR-CALNAME), if it has one.
	Name   string
	Events []Event
}

// Event is one VEVENT.
type Event struct {
	UID         string
	Summary     string
	Description string
	Location    string
	// Start and End are instants for timed events; for all-day events they
	// are dates at midnight UTC, and End is the day after the last day.
	Start, End time.Time
	AllDay     bool
	Rule       *Rule
	ExDates    []time.Time
	// RecurrenceID is set on a changed occurrence of a repeating event: the
	// start of the occurrence it replaces.
	RecurrenceID time.Time
	Cancelled    bool
}

// Rule is how an event repeats (RRULE).
type Rule struct {
	Freq       string // DAILY, WEEKLY, MONTHLY or YEARLY
	Interval   int
	Count      int
	Until      time.Time
	ByDay      []WeekdayNum
	ByMonthDay []int
	ByMonth    []int
	// Other is set when the rule has parts this package doesn't follow
	// (BYSETPOS, BYWEEKNO, BYYEARDAY, hourly and finer repeats).
	Other bool
}

// WeekdayNum is a BYDAY value: a weekday, and for monthly and yearly rules
// which one in the month (1 is the first, -1 the last, 0 every one).
type WeekdayNum struct {
	N   int
	Day time.Weekday
}

// Parse reads an iCalendar file. Times without a time zone are read in loc.
func Parse(r io.Reader, loc *time.Location) (*Calendar, error) {
	lines, err := unfold(r)
	if err != nil {
		return nil, err
	}
	cal := &Calendar{}
	var stack []string
	var ev *Event
	seen := false
	for _, raw := range lines {
		p, ok := parseLine(raw)
		if !ok {
			continue
		}
		switch p.name {
		case "BEGIN":
			c := strings.ToUpper(p.value)
			stack = append(stack, c)
			if c == "VCALENDAR" {
				seen = true
			}
			if c == "VEVENT" && len(stack) >= 2 && stack[len(stack)-2] == "VCALENDAR" {
				ev = &Event{}
			}
			continue
		case "END":
			c := strings.ToUpper(p.value)
			if c == "VEVENT" && ev != nil && len(stack) >= 1 && stack[len(stack)-1] == "VEVENT" {
				if !ev.Start.IsZero() {
					finish(ev)
					cal.Events = append(cal.Events, *ev)
				}
				ev = nil
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		if len(stack) == 0 {
			continue
		}
		top := stack[len(stack)-1]
		if top == "VCALENDAR" && p.name == "X-WR-CALNAME" {
			cal.Name = unescape(p.value)
		}
		if top != "VEVENT" || ev == nil {
			continue
		}
		switch p.name {
		case "UID":
			ev.UID = strings.TrimSpace(p.value)
		case "SUMMARY":
			ev.Summary = strings.TrimSpace(unescape(p.value))
		case "DESCRIPTION":
			ev.Description = strings.TrimSpace(unescape(p.value))
		case "LOCATION":
			ev.Location = strings.TrimSpace(unescape(p.value))
		case "STATUS":
			ev.Cancelled = strings.EqualFold(strings.TrimSpace(p.value), "CANCELLED")
		case "DTSTART":
			if t, allDay, err := parseTime(p, loc); err == nil {
				ev.Start, ev.AllDay = t, allDay
			}
		case "DTEND":
			if t, _, err := parseTime(p, loc); err == nil {
				ev.End = t
			}
		case "DURATION":
			if d, err := parseDuration(p.value); err == nil && !ev.Start.IsZero() {
				ev.End = ev.Start.Add(d)
			}
		case "RECURRENCE-ID":
			if t, _, err := parseTime(p, loc); err == nil {
				ev.RecurrenceID = t
			}
		case "RRULE":
			if rule, err := parseRule(p.value, loc); err == nil {
				ev.Rule = rule
			}
		case "EXDATE":
			for _, v := range strings.Split(p.value, ",") {
				q := p
				q.value = v
				if t, _, err := parseTime(q, loc); err == nil {
					ev.ExDates = append(ev.ExDates, t)
				}
			}
		}
	}
	if !seen {
		return nil, fmt.Errorf("this isn't a calendar file (.ics)")
	}
	return cal, nil
}

// finish fills in what an event left out.
func finish(ev *Event) {
	if ev.End.IsZero() || ev.End.Before(ev.Start) {
		if ev.AllDay {
			ev.End = ev.Start.AddDate(0, 0, 1)
		} else {
			ev.End = ev.Start
		}
	}
	if ev.AllDay && !ev.End.After(ev.Start) {
		ev.End = ev.Start.AddDate(0, 0, 1)
	}
}

// unfold reads the content lines, joining the ones folded onto the next line.
func unfold(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var lines []string
	for sc.Scan() {
		l := strings.TrimRight(sc.Text(), "\r")
		if l == "" {
			continue
		}
		if (l[0] == ' ' || l[0] == '\t') && len(lines) > 0 {
			lines[len(lines)-1] += l[1:]
			continue
		}
		lines = append(lines, l)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("couldn't read the calendar: %w", err)
	}
	if len(lines) > 0 {
		lines[0] = strings.TrimPrefix(lines[0], "\uFEFF")
	}
	return lines, nil
}

type property struct {
	name   string
	params map[string]string
	value  string
}

// parseLine splits "NAME;PARAM=x;PARAM="y:z":value".
func parseLine(l string) (property, bool) {
	inQuote := false
	colon := -1
	for i, c := range l {
		if c == '"' {
			inQuote = !inQuote
		} else if c == ':' && !inQuote {
			colon = i
			break
		}
	}
	if colon < 0 {
		return property{}, false
	}
	head, value := l[:colon], l[colon+1:]
	parts := splitParams(head)
	p := property{name: strings.ToUpper(parts[0]), value: value, params: map[string]string{}}
	for _, kv := range parts[1:] {
		k, v, _ := strings.Cut(kv, "=")
		p.params[strings.ToUpper(k)] = strings.Trim(v, `"`)
	}
	return p, true
}

func splitParams(s string) []string {
	var out []string
	inQuote := false
	start := 0
	for i, c := range s {
		if c == '"' {
			inQuote = !inQuote
		} else if c == ';' && !inQuote {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

func unescape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n', 'N':
				b.WriteByte('\n')
			default:
				b.WriteByte(s[i])
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// parseTime reads a DATE or DATE-TIME value. All-day dates come back at
// midnight UTC.
func parseTime(p property, loc *time.Location) (time.Time, bool, error) {
	v := strings.TrimSpace(p.value)
	if strings.EqualFold(p.params["VALUE"], "DATE") || len(v) == 8 {
		t, err := time.Parse("20060102", v[:min(8, len(v))])
		return t, true, err
	}
	if strings.HasSuffix(v, "Z") {
		t, err := time.Parse("20060102T150405Z", v)
		return t, false, err
	}
	in := loc
	if tz := p.params["TZID"]; tz != "" {
		in = Location(tz, loc)
	}
	t, err := time.ParseInLocation("20060102T150405", v, in)
	return t, false, err
}

// parseDuration reads a DURATION such as PT1H30M, P1D or P2W.
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimLeft(s, "+-")
	if !strings.HasPrefix(s, "P") {
		return 0, fmt.Errorf("bad duration %q", s)
	}
	var d time.Duration
	num := ""
	inTime := false
	for _, c := range s[1:] {
		switch {
		case c >= '0' && c <= '9':
			num += string(c)
		case c == 'T':
			inTime = true
		default:
			n, err := strconv.Atoi(num)
			if err != nil {
				return 0, fmt.Errorf("bad duration %q", s)
			}
			num = ""
			switch {
			case c == 'W':
				d += time.Duration(n) * 7 * 24 * time.Hour
			case c == 'D':
				d += time.Duration(n) * 24 * time.Hour
			case c == 'H' && inTime:
				d += time.Duration(n) * time.Hour
			case c == 'M' && inTime:
				d += time.Duration(n) * time.Minute
			case c == 'S' && inTime:
				d += time.Duration(n) * time.Second
			default:
				return 0, fmt.Errorf("bad duration %q", s)
			}
		}
	}
	if neg {
		d = -d
	}
	return d, nil
}

var weekdays = map[string]time.Weekday{
	"SU": time.Sunday, "MO": time.Monday, "TU": time.Tuesday, "WE": time.Wednesday,
	"TH": time.Thursday, "FR": time.Friday, "SA": time.Saturday,
}

func parseRule(s string, loc *time.Location) (*Rule, error) {
	r := &Rule{Interval: 1}
	for _, part := range strings.Split(s, ";") {
		k, v, _ := strings.Cut(part, "=")
		switch strings.ToUpper(k) {
		case "FREQ":
			r.Freq = strings.ToUpper(v)
		case "INTERVAL":
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				r.Interval = n
			}
		case "COUNT":
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				r.Count = n
			}
		case "UNTIL":
			if t, _, err := parseTime(property{value: v}, loc); err == nil {
				r.Until = t
			}
		case "BYDAY":
			for _, d := range strings.Split(v, ",") {
				d = strings.ToUpper(strings.TrimSpace(d))
				if len(d) < 2 {
					continue
				}
				wd, ok := weekdays[d[len(d)-2:]]
				if !ok {
					continue
				}
				n := 0
				if num := d[:len(d)-2]; num != "" {
					n, _ = strconv.Atoi(strings.TrimPrefix(num, "+"))
				}
				r.ByDay = append(r.ByDay, WeekdayNum{N: n, Day: wd})
			}
		case "BYMONTHDAY":
			r.ByMonthDay = ints(v)
		case "BYMONTH":
			r.ByMonth = ints(v)
		case "WKST", "":
		default:
			r.Other = true
		}
	}
	switch r.Freq {
	case "DAILY", "WEEKLY", "MONTHLY", "YEARLY":
	default:
		r.Other = true
	}
	return r, nil
}

func ints(s string) []int {
	var out []int
	for _, x := range strings.Split(s, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(x)); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// Location finds a TZID's time zone: an IANA name ("Europe/Brussels"), one
// with a prefix some apps add ("/mozilla.org/20050126_1/Europe/Brussels"), or
// a Windows name that Outlook uses ("W. Europe Standard Time"). Anything else
// is read in fallback.
func Location(tzid string, fallback *time.Location) *time.Location {
	tzid = strings.TrimSpace(tzid)
	if name, ok := windowsZones[tzid]; ok {
		tzid = name
	}
	parts := strings.Split(strings.Trim(tzid, "/"), "/")
	for i := range parts {
		name := strings.Join(parts[i:], "/")
		if name == "" || name == "Local" {
			continue
		}
		if l, err := time.LoadLocation(name); err == nil {
			return l
		}
	}
	return fallback
}

// The Windows time zone names Outlook and Exchange write most often.
var windowsZones = map[string]string{
	"UTC":                             "UTC",
	"GMT Standard Time":               "Europe/London",
	"Greenwich Standard Time":         "Atlantic/Reykjavik",
	"W. Europe Standard Time":         "Europe/Berlin",
	"Romance Standard Time":           "Europe/Paris",
	"Central Europe Standard Time":    "Europe/Budapest",
	"Central European Standard Time":  "Europe/Warsaw",
	"E. Europe Standard Time":         "Europe/Chisinau",
	"GTB Standard Time":               "Europe/Bucharest",
	"FLE Standard Time":               "Europe/Kiev",
	"Turkey Standard Time":            "Europe/Istanbul",
	"Russian Standard Time":           "Europe/Moscow",
	"Israel Standard Time":            "Asia/Jerusalem",
	"South Africa Standard Time":      "Africa/Johannesburg",
	"Arabian Standard Time":           "Asia/Dubai",
	"India Standard Time":             "Asia/Kolkata",
	"China Standard Time":             "Asia/Shanghai",
	"Singapore Standard Time":         "Asia/Singapore",
	"Tokyo Standard Time":             "Asia/Tokyo",
	"Korea Standard Time":             "Asia/Seoul",
	"AUS Eastern Standard Time":       "Australia/Sydney",
	"New Zealand Standard Time":       "Pacific/Auckland",
	"Eastern Standard Time":           "America/New_York",
	"Central Standard Time":           "America/Chicago",
	"Mountain Standard Time":          "America/Denver",
	"US Mountain Standard Time":       "America/Phoenix",
	"Pacific Standard Time":           "America/Los_Angeles",
	"Alaskan Standard Time":           "America/Anchorage",
	"Hawaiian Standard Time":          "Pacific/Honolulu",
	"Atlantic Standard Time":          "America/Halifax",
	"E. South America Standard Time":  "America/Sao_Paulo",
	"Argentina Standard Time":         "America/Buenos_Aires",
	"Canada Central Standard Time":    "America/Regina",
	"Central America Standard Time":   "America/Guatemala",
	"Mexico Standard Time":            "America/Mexico_City",
	"Central Standard Time (Mexico)":  "America/Mexico_City",
	"SA Pacific Standard Time":        "America/Bogota",
	"Egypt Standard Time":             "Africa/Cairo",
	"W. Central Africa Standard Time": "Africa/Lagos",
	"E. Africa Standard Time":         "Africa/Nairobi",
	"Morocco Standard Time":           "Africa/Casablanca",
	"Pakistan Standard Time":          "Asia/Karachi",
	"SE Asia Standard Time":           "Asia/Bangkok",
	"Taipei Standard Time":            "Asia/Taipei",
	"W. Australia Standard Time":      "Australia/Perth",
	"Cen. Australia Standard Time":    "Australia/Adelaide",
	"E. Australia Standard Time":      "Australia/Brisbane",
}
