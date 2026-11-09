import type { RecordModel } from 'pocketbase';
import { dateOf, dayOf, parseDate, type Day } from '../days';

// Daily habits: small things someone wants to do on some days ("brush teeth",
// twice a day; "gym", on Monday and Thursday) and how many times they did
// them each day.

export type Weekday = 'mon' | 'tue' | 'wed' | 'thu' | 'fri' | 'sat' | 'sun';
export type Part = 'morning' | 'afternoon' | 'evening';

export interface Habit extends RecordModel {
  user: string;
  title: string;
  icon: string;
  /** How many times a day: brushing teeth is 2. */
  times: number;
  /** The weekdays it's due on; none means every day. */
  days: Weekday[];
  part: Part | '';
  sort: number;
  archived: boolean;
  created: string;
}

/** What someone did about a habit on a day. There is at most one per habit and day. */
export interface Check extends RecordModel {
  habit: string;
  day: Day;
  count: number;
  /** Not today, on purpose: it isn't due that day. */
  skipped: boolean;
}

export const WEEKDAYS: { day: Weekday; label: string }[] = [
  { day: 'mon', label: 'Mo' },
  { day: 'tue', label: 'Tu' },
  { day: 'wed', label: 'We' },
  { day: 'thu', label: 'Th' },
  { day: 'fri', label: 'Fr' },
  { day: 'sat', label: 'Sa' },
  { day: 'sun', label: 'Su' },
];

export const PARTS: { part: Part | ''; label: string }[] = [
  { part: 'morning', label: 'Morning' },
  { part: 'afternoon', label: 'Afternoon' },
  { part: 'evening', label: 'Evening' },
  { part: '', label: 'Any time' },
];

/** The icons a habit can have (see Icon.svelte). */
export const HABIT_ICONS = [
  'tooth',
  'pill',
  'food',
  'drop',
  'plant',
  'ball',
  'book',
  'moon',
  'sun',
  'heart',
  'music',
  'home',
] as const;

/** Habits to start with, picked on the first visit. */
export const STARTERS: Pick<Habit, 'title' | 'icon' | 'times' | 'part'>[] = [
  { title: 'Brush teeth', icon: 'tooth', times: 2, part: '' },
  { title: 'Vitamin D', icon: 'pill', times: 1, part: 'morning' },
  { title: 'Eat healthy', icon: 'food', times: 1, part: '' },
  { title: 'Drink water', icon: 'drop', times: 1, part: '' },
  { title: 'Go for a walk', icon: 'plant', times: 1, part: 'afternoon' },
  { title: 'Read a little', icon: 'book', times: 1, part: 'evening' },
];

const weekday = (day: Day): Weekday => WEEKDAYS[(dateOf(day).getDay() + 6) % 7].day;

/** The day a habit was added, from which on it counts. */
function since(h: Habit): Day {
  const d = parseDate(h.created);
  return d ? dayOf(d) : '0000-00-00';
}

export function checkOf(checks: Check[], h: Habit, day: Day): Check | undefined {
  return checks.find((c) => c.habit === h.id && c.day === day);
}

/** Whether a habit is something to do on a day: on its weekdays, not skipped. */
export function isDue(h: Habit, day: Day, checks: Check[]): boolean {
  if (h.archived || day < since(h)) return false;
  if (h.days?.length && !h.days.includes(weekday(day))) return false;
  return !checkOf(checks, h, day)?.skipped;
}

/** How many times it was done on a day, up to how many times it's due. */
export function doneCount(h: Habit, day: Day, checks: Check[]): number {
  return Math.min(checkOf(checks, h, day)?.count ?? 0, h.times);
}

export const isDone = (h: Habit, day: Day, checks: Check[]) => doneCount(h, day, checks) >= h.times;

/**
 * How much of a day's habits were done, from 0 to 1, with a habit done once of
 * twice counting half. Null when nothing was due.
 */
export function progress(habits: Habit[], checks: Check[], day: Day): number | null {
  const due = habits.filter((h) => isDue(h, day, checks));
  if (!due.length) return null;
  return due.reduce((sum, h) => sum + doneCount(h, day, checks) / h.times, 0) / due.length;
}

/** The part of the day it is now. */
export function partNow(now = new Date()): Part {
  const h = now.getHours();
  return h < 12 ? 'morning' : h < 18 ? 'afternoon' : 'evening';
}

/** "Every day", "Weekdays", "Mo, We, Fr"; with "twice a day" when it's more than once. */
export function scheduleLabel(h: Pick<Habit, 'days' | 'times'>): string {
  const days = h.days ?? [];
  const on =
    !days.length || days.length === 7
      ? 'Every day'
      : days.length === 5 && !days.includes('sat') && !days.includes('sun')
        ? 'Weekdays'
        : WEEKDAYS.filter((w) => days.includes(w.day))
            .map((w) => w.label)
            .join(', ');
  if (h.times === 1) return on;
  return `${on} · ${h.times === 2 ? 'twice' : `${h.times} times`} a day`;
}
