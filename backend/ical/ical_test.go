package ical

import (
	"strings"
	"testing"
	"time"
)

// Shaped like Google Calendar's export: folded lines, escapes, an all-day
// event, a timed one in a time zone, a weekly repeat with a changed time and
// a cancelled one.
const google = "BEGIN:VCALENDAR\r\n" +
	"PRODID:-//Google Inc//Google Calendar 70.9054//EN\r\n" +
	"VERSION:2.0\r\n" +
	"X-WR-CALNAME:Family\r\n" +
	"BEGIN:VTIMEZONE\r\nTZID:Europe/Brussels\r\nBEGIN:STANDARD\r\nDTSTART:19701025T030000\r\nEND:STANDARD\r\nEND:VTIMEZONE\r\n" +
	"BEGIN:VEVENT\r\n" +
	"DTSTART;VALUE=DATE:20261212\r\n" +
	"DTEND;VALUE=DATE:20261219\r\n" +
	"UID:trip@google.com\r\n" +
	"SUMMARY:Crete\\, with friends\r\n" +
	"DESCRIPTION:Flight at 7.\\nBring the \r\n  passports.\r\n" +
	"END:VEVENT\r\n" +
	"BEGIN:VEVENT\r\n" +
	"DTSTART;TZID=Europe/Brussels:20261205T183000\r\n" +
	"DTEND;TZID=Europe/Brussels:20261205T210000\r\n" +
	"UID:dinner@google.com\r\n" +
	"SUMMARY:Dinner at Maria's\r\n" +
	"LOCATION:Rue de la Loi 1\\, Brussels\r\n" +
	"BEGIN:VALARM\r\nACTION:DISPLAY\r\nDESCRIPTION:This is an event reminder\r\nEND:VALARM\r\n" +
	"END:VEVENT\r\n" +
	"BEGIN:VEVENT\r\n" +
	"DTSTART:20261202T153000Z\r\n" +
	"DURATION:PT1H\r\n" +
	"RRULE:FREQ=WEEKLY;BYDAY=WE;UNTIL=20270630T000000Z\r\n" +
	"EXDATE:20261223T153000Z\r\n" +
	"UID:piano@google.com\r\n" +
	"SUMMARY:Piano\r\n" +
	"END:VEVENT\r\n" +
	"BEGIN:VEVENT\r\n" +
	"DTSTART:20261209T163000Z\r\n" +
	"DTEND:20261209T173000Z\r\n" +
	"RECURRENCE-ID:20261209T153000Z\r\n" +
	"UID:piano@google.com\r\n" +
	"SUMMARY:Piano (later)\r\n" +
	"END:VEVENT\r\n" +
	"BEGIN:VEVENT\r\n" +
	"DTSTART;VALUE=DATE:20261220\r\n" +
	"UID:off@google.com\r\n" +
	"STATUS:CANCELLED\r\n" +
	"SUMMARY:Called off\r\n" +
	"END:VEVENT\r\n" +
	"END:VCALENDAR\r\n"

func TestParse(t *testing.T) {
	cal, err := Parse(strings.NewReader(google), time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if cal.Name != "Family" {
		t.Errorf("Name = %q", cal.Name)
	}
	if len(cal.Events) != 5 {
		t.Fatalf("got %d events", len(cal.Events))
	}

	trip := cal.Events[0]
	if trip.Summary != "Crete, with friends" || !trip.AllDay ||
		trip.Description != "Flight at 7.\nBring the  passports." {
		t.Errorf("trip = %+v", trip)
	}
	if got := trip.Start.Format(time.DateOnly) + " " + trip.End.Format(time.DateOnly); got != "2026-12-12 2026-12-19" {
		t.Errorf("trip dates = %s", got)
	}

	dinner := cal.Events[1]
	if dinner.Location != "Rue de la Loi 1, Brussels" {
		t.Errorf("location = %q (the alarm's description must not count)", dinner.Location)
	}
	if got := dinner.Start.UTC().Format(time.RFC3339); got != "2026-12-05T17:30:00Z" {
		t.Errorf("dinner starts %s", got)
	}

	piano := cal.Events[2]
	if piano.Rule == nil || piano.Rule.Freq != "WEEKLY" || len(piano.ExDates) != 1 {
		t.Fatalf("piano = %+v", piano)
	}
	if piano.End.Sub(piano.Start) != time.Hour {
		t.Errorf("DURATION not read: %v", piano.End.Sub(piano.Start))
	}
	if cal.Events[3].RecurrenceID.IsZero() || !cal.Events[4].Cancelled {
		t.Error("the changed and the cancelled events aren't marked")
	}
}

func TestNotACalendar(t *testing.T) {
	if _, err := Parse(strings.NewReader("<html>Sign in</html>"), time.UTC); err == nil {
		t.Error("an HTML page parsed as a calendar")
	}
}

func TestLocation(t *testing.T) {
	for tzid, want := range map[string]string{
		"Europe/Athens":                       "Europe/Athens",
		"W. Europe Standard Time":             "Europe/Berlin",
		"/mozilla.org/20050126_1/Europe/Rome": "Europe/Rome",
		"Made up":                             "UTC",
	} {
		if got := Location(tzid, time.UTC).String(); got != want {
			t.Errorf("Location(%q) = %s, want %s", tzid, got, want)
		}
	}
}

func days(ts []time.Time) string {
	var out []string
	for _, t := range ts {
		out = append(out, t.Format("2006-01-02"))
	}
	return strings.Join(out, " ")
}

func TestOccurrences(t *testing.T) {
	at := func(s string) time.Time {
		v, err := time.Parse(time.DateOnly, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	from, to := at("2026-01-01"), at("2027-01-01")
	scenarios := []struct {
		name, start, rule string
		exdates           []string
		want              string
	}{
		{"every other week", "2026-03-02", "FREQ=WEEKLY;INTERVAL=2;COUNT=4", nil, "2026-03-02 2026-03-16 2026-03-30 2026-04-13"},
		{"weekdays", "2026-03-05", "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;COUNT=4", nil, "2026-03-05 2026-03-06 2026-03-09 2026-03-10"},
		{"second Tuesday", "2026-01-13", "FREQ=MONTHLY;BYDAY=2TU;COUNT=3", nil, "2026-01-13 2026-02-10 2026-03-10"},
		{"last Friday", "2026-01-30", "FREQ=MONTHLY;BYDAY=-1FR;COUNT=3", nil, "2026-01-30 2026-02-27 2026-03-27"},
		{"the 31st skips short months", "2026-01-31", "FREQ=MONTHLY;COUNT=3", nil, "2026-01-31 2026-03-31 2026-05-31"},
		{"until, with a day skipped", "2026-06-01", "FREQ=DAILY;UNTIL=20260604", []string{"2026-06-02"}, "2026-06-01 2026-06-03 2026-06-04"},
		{"yearly in the window", "2020-05-04", "FREQ=YEARLY", nil, "2026-05-04"},
	}
	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			rule, _ := parseRule(s.rule, time.UTC)
			ev := Event{Start: at(s.start), End: at(s.start).AddDate(0, 0, 1), AllDay: true, Rule: rule}
			for _, x := range s.exdates {
				ev.ExDates = append(ev.ExDates, at(x))
			}
			if got := days(ev.Occurrences(from, to, 100)); got != s.want {
				t.Errorf("got  %s\nwant %s", got, s.want)
			}
		})
	}
}
