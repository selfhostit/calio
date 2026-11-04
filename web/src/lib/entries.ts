import type { RecordModel } from 'pocketbase';

// Everything on the household calendar. All-day entries keep their dates at
// midnight UTC and are read as plain calendar dates; timed entries are instants.

export type Kind = 'event' | 'petsit' | 'trip';
export type SitterStatus = 'needed' | 'asked' | 'booked';

export interface Entry extends RecordModel {
  title: string;
  kind: Kind;
  start: string;
  end: string;
  all_day: boolean;
  location: string;
  notes: string;
  people: string[];
  sitter: string;
  sitter_status: SitterStatus | '';
  created_by: string;
}

export const KINDS: { kind: Kind; label: string; plural: string }[] = [
  { kind: 'event', label: 'Event', plural: 'Events' },
  { kind: 'petsit', label: 'Pet sitting', plural: 'Pet sitting' },
  { kind: 'trip', label: 'Trip', plural: 'Trips' },
];

export const SITTER_STATUS: { status: SitterStatus; label: string }[] = [
  { status: 'needed', label: 'Needed' },
  { status: 'asked', label: 'Asked' },
  { status: 'booked', label: 'Booked' },
];

/** A calendar day as YYYY-MM-DD; these sort and compare as strings. */
export type Day = string;

const pad = (n: number) => String(n).padStart(2, '0');

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

/** Parses a PocketBase date ("2026-12-12 18:00:00.000Z"). */
export function parseDate(s: string): Date | null {
  return s ? new Date(s.replace(' ', 'T')) : null;
}

/** The first and last day an entry is on. */
export function span(e: Pick<Entry, 'start' | 'end' | 'all_day'>): [Day, Day] {
  if (e.all_day) {
    const first = e.start.slice(0, 10);
    const last = e.end ? e.end.slice(0, 10) : first;
    return [first, last < first ? first : last];
  }
  const start = parseDate(e.start)!;
  const end = parseDate(e.end) ?? start;
  const first = dayOf(start);
  // a party that ends at midnight is still only on its first day
  const last = dayOf(new Date(Math.max(start.getTime(), end.getTime() - 1)));
  return [first, last < first ? first : last];
}

export function isOn(e: Entry, day: Day): boolean {
  const [first, last] = span(e);
  return first <= day && day <= last;
}

/** Sorts entries by when they start, all-day ones first on the same day. */
export function byStart(a: Entry, b: Entry): number {
  const [da] = span(a);
  const [db] = span(b);
  if (da !== db) return da < db ? -1 : 1;
  if (a.all_day !== b.all_day) return a.all_day ? -1 : 1;
  return a.start < b.start ? -1 : a.start > b.start ? 1 : 0;
}

/**
 * Groups what's coming up by day: each entry goes under the day it starts, or
 * under today when it has already started and is still going.
 */
export function agenda(entries: Entry[], today: Day): { day: Day; entries: Entry[] }[] {
  const groups = new Map<Day, Entry[]>();
  for (const e of [...entries].sort(byStart)) {
    const [first, last] = span(e);
    if (last < today) continue;
    const day = first < today ? today : first;
    groups.set(day, [...(groups.get(day) ?? []), e]);
  }
  return [...groups]
    .map(([day, entries]) => ({ day, entries }))
    .sort((a, b) => (a.day < b.day ? -1 : 1));
}

/** The six weeks shown for a month, starting on Monday. */
export function monthGrid(year: number, month: number): Day[] {
  const first = new Date(year, month, 1);
  const offset = (first.getDay() + 6) % 7;
  const start = dayOf(new Date(year, month, 1 - offset));
  return Array.from({ length: 42 }, (_, i) => addDays(start, i));
}

// ---- Words for dates ----

const fmt = (opts: Intl.DateTimeFormatOptions) => new Intl.DateTimeFormat(undefined, opts);
const weekdayDay = fmt({ weekday: 'short', day: 'numeric', month: 'short' });
const longDay = fmt({ weekday: 'long', day: 'numeric', month: 'long' });
const time = fmt({ hour: 'numeric', minute: '2-digit' });
const range = fmt({ day: 'numeric', month: 'short' });
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

/** When an entry happens, like "18:30 – 21:00", "All day" or "12 – 18 Dec". */
export function whenLabel(e: Entry): string {
  const [first, last] = span(e);
  if (first !== last) {
    const a = dateOf(first);
    const b = dateOf(last);
    const nights = Math.round((b.getTime() - a.getTime()) / 864e5);
    const timed = e.all_day ? '' : ` · ${time.format(parseDate(e.start)!)}`;
    return `${range.formatRange(a, b)} · ${nights} ${nights === 1 ? 'night' : 'nights'}${timed}`;
  }
  if (e.all_day) return 'All day';
  const start = time.format(parseDate(e.start)!);
  return e.end ? `${start} – ${time.format(parseDate(e.end)!)}` : start;
}

/** For an entry that is going on today: "Day 3 of 7". */
export function progressLabel(e: Entry, today: Day): string {
  const [first, last] = span(e);
  if (first === last || today < first || today > last) return '';
  const day = Math.round((dateOf(today).getTime() - dateOf(first).getTime()) / 864e5) + 1;
  const total = Math.round((dateOf(last).getTime() - dateOf(first).getTime()) / 864e5) + 1;
  return `Day ${day} of ${total}`;
}

/** "in 3 days", "tomorrow", "today". */
export function inDays(day: Day, today: Day): string {
  const n = Math.round((dateOf(day).getTime() - dateOf(today).getTime()) / 864e5);
  if (n <= 0) return 'today';
  if (n === 1) return 'tomorrow';
  return `in ${n} days`;
}

// ---- Between the form and the database ----

/** A Day as an all-day value for the database. */
export const allDayValue = (day: Day) => `${day} 00:00:00.000Z`;

/** A datetime-local input value ("2026-12-12T18:30") as a value for the database. */
export const instantValue = (local: string) =>
  local ? new Date(local).toISOString().replace('T', ' ') : '';

/** A database date as a datetime-local input value. */
export function localInputValue(s: string): string {
  const d = parseDate(s);
  if (!d) return '';
  return `${dayOf(d)}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
