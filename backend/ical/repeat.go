package ical

import (
	"sort"
	"time"
)

// Occurrences lists the starts of a repeating event's occurrences that begin
// before `to` and end after `from`, at most max of them. It follows the rule's
// frequency, interval, BYDAY, BYMONTHDAY, BYMONTH, COUNT and UNTIL, and leaves
// out the EXDATEs. An event that doesn't repeat has one occurrence.
func (e Event) Occurrences(from, to time.Time, max int) []time.Time {
	length := e.End.Sub(e.Start)
	in := func(t time.Time) bool { return t.Before(to) && !t.Add(length).Before(from) }
	if e.Rule == nil {
		if in(e.Start) {
			return []time.Time{e.Start}
		}
		return nil
	}
	excluded := map[int64]bool{}
	for _, x := range e.ExDates {
		excluded[x.Unix()] = true
		if e.AllDay {
			// an EXDATE given as a date-time still means that day
			excluded[time.Date(x.Year(), x.Month(), x.Day(), 0, 0, 0, 0, time.UTC).Unix()] = true
		}
	}

	r := e.Rule
	var out []time.Time
	n := 0
	// periods to step through: enough for many years of daily repeats
	for period := 0; period < 50000; period++ {
		var candidates []time.Time
		for _, t := range r.period(e.Start, period) {
			if !t.Before(e.Start) {
				candidates = append(candidates, t)
			}
		}
		for _, t := range candidates {
			if !r.Until.IsZero() && t.After(untilFor(r.Until, e)) {
				return out
			}
			if !t.Before(to) {
				return out
			}
			n++
			if r.Count > 0 && n > r.Count {
				return out
			}
			if excluded[t.Unix()] || !in(t) {
				continue
			}
			out = append(out, t)
			if len(out) >= max {
				return out
			}
		}
	}
	return out
}

// untilFor reads UNTIL for an all-day event as the end of that day.
func untilFor(until time.Time, e Event) time.Time {
	if e.AllDay {
		return time.Date(until.Year(), until.Month(), until.Day(), 0, 0, 0, 0, time.UTC)
	}
	return until
}

// period gives the candidate starts in the nth period (day, week, month or
// year, times the interval) after the start, in order.
func (r *Rule) period(start time.Time, n int) []time.Time {
	loc := start.Location()
	y, m, d := start.Date()
	hh, mm, ss := start.Clock()
	at := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, hh, mm, ss, 0, loc) }
	step := n * r.Interval
	var out []time.Time
	switch r.Freq {
	case "DAILY":
		t := at(y, m, d+step)
		if r.monthOK(t.Month()) && r.dayOK(t.Weekday()) {
			out = append(out, t)
		}
	case "WEEKLY":
		// weeks start on Monday
		monday := at(y, m, d-(int(start.Weekday())+6)%7+7*step)
		if len(r.ByDay) == 0 {
			out = append(out, at(y, m, d+7*step))
			break
		}
		for i := 0; i < 7; i++ {
			t := monday.AddDate(0, 0, i)
			for _, wd := range r.ByDay {
				if wd.Day == t.Weekday() {
					out = append(out, at(t.Year(), t.Month(), t.Day()))
				}
			}
		}
	case "MONTHLY":
		first := time.Date(y, m+time.Month(step), 1, 0, 0, 0, 0, loc)
		out = r.inMonth(first.Year(), first.Month(), d, at)
	case "YEARLY":
		year := y + step
		months := r.ByMonth
		if len(months) == 0 {
			months = []int{int(m)}
		}
		for _, mo := range months {
			if len(r.ByDay) == 0 && len(r.ByMonthDay) == 0 {
				// the same day each year; skipped in years without it (29 February)
				if t := at(year, time.Month(mo), d); t.Day() == d {
					out = append(out, t)
				}
				continue
			}
			out = append(out, r.inMonth(year, time.Month(mo), d, at)...)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	return out
}

func (r *Rule) dayOK(wd time.Weekday) bool {
	if len(r.ByDay) == 0 {
		return true
	}
	for _, x := range r.ByDay {
		if x.Day == wd {
			return true
		}
	}
	return false
}

func (r *Rule) monthOK(m time.Month) bool {
	if len(r.ByMonth) == 0 {
		return true
	}
	for _, x := range r.ByMonth {
		if time.Month(x) == m {
			return true
		}
	}
	return false
}

// inMonth gives the days of one month the rule picks: by BYMONTHDAY, by BYDAY
// ("the second Tuesday", "the last Friday", "every Monday") or the start's day.
func (r *Rule) inMonth(y int, m time.Month, day int, at func(int, time.Month, int) time.Time) []time.Time {
	days := daysIn(y, m)
	var out []time.Time
	switch {
	case len(r.ByMonthDay) > 0:
		for _, md := range r.ByMonthDay {
			if md < 0 {
				md = days + md + 1
			}
			if md >= 1 && md <= days {
				out = append(out, at(y, m, md))
			}
		}
	case len(r.ByDay) > 0:
		for _, wd := range r.ByDay {
			var matches []int
			for dd := 1; dd <= days; dd++ {
				if time.Date(y, m, dd, 12, 0, 0, 0, time.UTC).Weekday() == wd.Day {
					matches = append(matches, dd)
				}
			}
			switch {
			case wd.N == 0:
				for _, dd := range matches {
					out = append(out, at(y, m, dd))
				}
			case wd.N > 0 && wd.N <= len(matches):
				out = append(out, at(y, m, matches[wd.N-1]))
			case wd.N < 0 && -wd.N <= len(matches):
				out = append(out, at(y, m, matches[len(matches)+wd.N]))
			}
		}
	default:
		// months without that day (the 31st) are skipped
		if day <= days {
			out = append(out, at(y, m, day))
		}
	}
	return out
}

func daysIn(y int, m time.Month) int {
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
