import type { RecordModel } from 'pocketbase';
import { addDays, dateOf, dayOf, daysBetween, fmt, pad, parseDate, type Day } from './days';

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
  repeat: Repeat | '';
  repeat_until: string;
  checklist: ChecklistItem[] | null;
  /** A list of things to buy, which can be sent to Shoppa. */
  shopping: ChecklistItem[] | null;
  /** The category's id; the kind follows it. */
  category: string;
  /** The calendar it comes from, when it was subscribed to by its link. */
  calendar?: string;
  /** Its id in the calendar it was imported from. */
  uid?: string;
  /** Set on a repeat of an entry: a key unique to that occurrence. */
  key?: string;
}

export type Repeat = 'daily' | 'weekly' | 'monthly' | 'yearly';

export interface ChecklistItem {
  text: string;
  done: boolean;
}

export const REPEATS: { repeat: Repeat | ''; label: string }[] = [
  { repeat: '', label: 'Never' },
  { repeat: 'daily', label: 'Every day' },
  { repeat: 'weekly', label: 'Every week' },
  { repeat: 'monthly', label: 'Every month' },
  { repeat: 'yearly', label: 'Every year' },
];

export const SITTER_STATUS: { status: SitterStatus; label: string }[] = [
  { status: 'needed', label: 'Needed' },
  { status: 'asked', label: 'Asked' },
  { status: 'booked', label: 'Booked' },
];

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

/**
 * The day of the nth repeat after `first`, or null for a month or year that
 * doesn't have that day (the 31st, or 29 February), which is skipped.
 */
export function nthRepeat(first: Day, repeat: Repeat, n: number): Day | null {
  const [y, m, d] = first.split('-').map(Number);
  if (repeat === 'daily') return addDays(first, n);
  if (repeat === 'weekly') return addDays(first, 7 * n);
  const date = repeat === 'monthly' ? new Date(y, m - 1 + n, d) : new Date(y + n, m - 1, d);
  return date.getDate() === d ? dayOf(date) : null;
}

/**
 * Entries as they happen between two days: a repeating entry becomes one copy
 * per repeat, moved to its day and keyed by it.
 */
export function occurrences(entries: Entry[], from: Day, to: Day): Entry[] {
  const out: Entry[] = [];
  for (const e of entries) {
    if (!e.repeat) {
      out.push(e);
      continue;
    }
    const [first, last] = span(e);
    const length = daysBetween(first, last);
    const until = e.repeat_until ? e.repeat_until.slice(0, 10) : '9999-12-31';
    for (let n = 0; n < 100_000; n++) {
      const day = n === 0 ? first : nthRepeat(first, e.repeat, n);
      if (day === null) continue;
      if (day > to || day > until) break;
      if (addDays(day, length) >= from) out.push(n === 0 ? { ...e, key: e.id } : moveTo(e, day, n));
    }
  }
  return out;
}

/** A copy of an entry moved to another first day, keeping its times. */
function moveTo(e: Entry, day: Day, n: number): Entry {
  const [first] = span(e);
  const shift = (s: string) => {
    if (!s) return s;
    if (e.all_day) return allDayValue(addDays(s.slice(0, 10), daysBetween(first, day)));
    const d = parseDate(s)!;
    const moved = dateOf(addDays(dayOf(d), daysBetween(first, day)));
    moved.setHours(d.getHours(), d.getMinutes(), d.getSeconds());
    return moved.toISOString().replace('T', ' ');
  };
  return { ...e, start: shift(e.start), end: shift(e.end), key: `${e.id}:${n}` };
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

// ---- Words for entries ----

const time = fmt({ hour: 'numeric', minute: '2-digit' });
const range = fmt({ day: 'numeric', month: 'short' });

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
  const day = daysBetween(first, today) + 1;
  const total = daysBetween(first, last) + 1;
  return `Day ${day} of ${total}`;
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

// ---- Places ----

/**
 * A link that opens a location in Google Maps. A pasted map link is used as
 * it is; an address or "lat, lon" is searched for.
 */
export function mapLink(location: string): string {
  const l = location.trim();
  if (/^https?:\/\//i.test(l)) return l;
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(l)}`;
}

/** What to show for a location: a pasted link reads as "Map". */
export function placeLabel(location: string): string {
  return /^https?:\/\//i.test(location.trim()) ? 'Map' : location;
}

/** "3/12" for a checklist (or a shopping list), or '' when there is none. */
export function checklistLabel(e: Pick<Entry, 'checklist'>): string {
  return listLabel(e.checklist);
}

export function listLabel(list: ChecklistItem[] | null): string {
  const items = list ?? [];
  if (!items.length) return '';
  return `${items.filter((i) => i.done).length}/${items.length}`;
}
