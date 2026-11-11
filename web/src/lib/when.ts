import { addDays, daysBetween, dayOf, pad, parseDate, type Day } from './days';
import { allDayValue, instantValue, localInputValue, span, type Entry } from './entries';

// When an entry happens, the way the form edits it: plain days and times of
// day. Moving the start moves the end with it, so a trip keeps its nights and
// a dinner its hours, as in other calendar apps.

export interface When {
  allDay: boolean;
  startDay: Day;
  /** The last day; the same as the start for an entry on one day. */
  endDay: Day;
  /** "18:30"; only used when it isn't all day. */
  startTime: string;
  /** '' for an entry without an end time. */
  endTime: string;
}

export type Dates = Pick<Entry, 'all_day' | 'start' | 'end'>;

const DEFAULT_TIME = '18:00';

/** A new entry on a day. */
export function whenOn(day: Day, allDay: boolean): When {
  return { allDay, startDay: day, endDay: day, startTime: DEFAULT_TIME, endTime: '' };
}

/** When an entry that exists happens. */
export function whenOf(e: Dates): When {
  if (e.all_day) {
    const [startDay, endDay] = span(e);
    return { allDay: true, startDay, endDay, startTime: DEFAULT_TIME, endTime: '' };
  }
  const [startDay, startTime] = localInputValue(e.start).split('T');
  const end = localInputValue(e.end);
  const [endDay, endTime] = end ? end.split('T') : [startDay, ''];
  return { allDay: false, startDay, endDay, startTime, endTime };
}

/** The entry's dates as they go in the database. */
export function datesOf(w: When): Dates {
  if (w.allDay) {
    return {
      all_day: true,
      start: allDayValue(w.startDay),
      end: w.endDay > w.startDay ? allDayValue(w.endDay) : '',
    };
  }
  return {
    all_day: false,
    start: instantValue(`${w.startDay}T${w.startTime}`),
    end: w.endTime ? instantValue(`${w.endDay}T${w.endTime}`) : '',
  };
}

const at = (day: Day, time: string) => new Date(`${day}T${time}`);
const timeOf = (d: Date) => `${pad(d.getHours())}:${pad(d.getMinutes())}`;

/** A new start; the end moves with it, keeping how long it lasts. */
export function withStart(w: When, startDay: Day, startTime = w.startTime): When {
  if (w.allDay || !w.endTime) {
    return {
      ...w,
      startDay,
      startTime,
      endDay: addDays(w.endDay, daysBetween(w.startDay, startDay)),
    };
  }
  const length = at(w.endDay, w.endTime).getTime() - at(w.startDay, w.startTime).getTime();
  const end = new Date(at(startDay, startTime).getTime() + Math.max(0, length));
  return { ...w, startDay, startTime, endDay: dayOf(end), endTime: timeOf(end) };
}

/** Moves the whole thing by a number of days. */
export const shifted = (w: When, days: number) => withStart(w, addDays(w.startDay, days));

/**
 * All day or not. A timed entry that runs over several days keeps its days,
 * from and until the same time.
 */
export function withAllDay(w: When, allDay: boolean): When {
  if (allDay === w.allDay) return w;
  if (allDay) return { ...w, allDay };
  return { ...w, allDay, endTime: w.endDay > w.startDay ? w.startTime : '' };
}

/** What's wrong with it, or ''. */
export function problem(w: When): string {
  if (w.allDay) return w.endDay < w.startDay ? 'The last day can’t be before the first.' : '';
  if (!w.startTime) return 'Pick a start time.';
  if (w.endTime && at(w.endDay, w.endTime) < at(w.startDay, w.startTime))
    return 'The end can’t be before the start.';
  return '';
}

/** How long it lasts, like "3 nights" or "1 h 30 min", or '' for one day or no end. */
export function lengthLabel(w: When): string {
  if (w.allDay) {
    const nights = daysBetween(w.startDay, w.endDay);
    if (nights <= 0) return '';
    return `${nights + 1} days · ${nights} ${nights === 1 ? 'night' : 'nights'}`;
  }
  if (!w.endTime) return '';
  const minutes = Math.round(
    (at(w.endDay, w.endTime).getTime() - at(w.startDay, w.startTime).getTime()) / 60_000,
  );
  if (minutes <= 0) return '';
  const days = Math.floor(minutes / 1440);
  const h = Math.floor((minutes % 1440) / 60);
  const m = minutes % 60;
  return [days && `${days} d`, h && `${h} h`, m && `${m} min`].filter(Boolean).join(' ');
}

/** Whether two sets of dates are the same. */
export function sameDates(a: Dates, b: Dates): boolean {
  const t = (s: string) => parseDate(s)?.getTime() ?? 0;
  return a.all_day === b.all_day && t(a.start) === t(b.start) && t(a.end) === t(b.end);
}
