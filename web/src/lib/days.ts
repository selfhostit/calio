// Calendar days, shared by everything that works by the day: the calendar,
// and anything else that happens "today". A Day is a plain local date, not an
// instant.

/** A calendar day as YYYY-MM-DD; these sort and compare as strings. */
export type Day = string;

/** Two digits: 7 as "07". */
export const pad = (n: number) => String(n).padStart(2, '0');

/** The local calendar day of a Date. */
export function dayOf(d: Date): Day {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** A Day as a local Date at midnight. */
export function dateOf(day: Day): Date {
  const [y, m, d] = day.split('-').map(Number);
  return new Date(y, m - 1, d);
}

export function addDays(day: Day, n: number): Day {
  const d = dateOf(day);
  d.setDate(d.getDate() + n);
  return dayOf(d);
}

/** How many days from a to b: 1 from today to tomorrow. */
export const daysBetween = (a: Day, b: Day) =>
  Math.round((dateOf(b).getTime() - dateOf(a).getTime()) / 864e5);

/** Parses a PocketBase date ("2026-12-12 18:00:00.000Z"). */
export function parseDate(s: string): Date | null {
  return s ? new Date(s.replace(' ', 'T')) : null;
}

/** The six weeks shown for a month, starting on Monday. */
export function monthGrid(year: number, month: number): Day[] {
  const first = new Date(year, month, 1);
  const offset = (first.getDay() + 6) % 7;
  const start = dayOf(new Date(year, month, 1 - offset));
  return Array.from({ length: 42 }, (_, i) => addDays(start, i));
}

// ---- Words for days ----

export const fmt = (opts: Intl.DateTimeFormatOptions) => new Intl.DateTimeFormat(undefined, opts);
const weekdayDay = fmt({ weekday: 'short', day: 'numeric', month: 'short' });
const longDay = fmt({ weekday: 'long', day: 'numeric', month: 'long' });
export const monthName = fmt({ month: 'long', year: 'numeric' });

/** "Today", "Tomorrow", or "Sat, 12 Dec". */
export function dayLabel(day: Day, today: Day): string {
  if (day === today) return 'Today';
  if (day === addDays(today, 1)) return 'Tomorrow';
  if (day === addDays(today, -1)) return 'Yesterday';
  return weekdayDay.format(dateOf(day));
}

export function longDayLabel(day: Day): string {
  return longDay.format(dateOf(day));
}

/** "in 3 days", "tomorrow", "today". */
export function inDays(day: Day, today: Day): string {
  const n = daysBetween(today, day);
  if (n <= 0) return 'today';
  if (n === 1) return 'tomorrow';
  return `in ${n} days`;
}
