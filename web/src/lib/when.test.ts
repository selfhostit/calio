import { describe, expect, it } from 'vitest';
import { allDayValue, instantValue, shiftBy } from './entries';
import {
  datesOf,
  lengthLabel,
  problem,
  sameDates,
  shifted,
  whenOf,
  whenOn,
  withAllDay,
  withStart,
  type When,
} from './when';

const trip: When = {
  allDay: true,
  startDay: '2026-12-12',
  endDay: '2026-12-18',
  startTime: '18:00',
  endTime: '',
};
const dinner: When = {
  allDay: false,
  startDay: '2026-12-05',
  endDay: '2026-12-05',
  startTime: '18:30',
  endTime: '21:00',
};

describe('when', () => {
  it('reads and writes the dates of an entry', () => {
    for (const w of [trip, dinner]) expect(whenOf(datesOf(w))).toEqual(w);
    expect(datesOf(whenOn('2026-12-12', true))).toEqual({
      all_day: true,
      start: allDayValue('2026-12-12'),
      end: '',
    });
    expect(datesOf({ ...dinner, endTime: '' }).end).toBe('');
  });

  it('keeps how long it lasts when the start moves', () => {
    expect(withStart(trip, '2026-12-20')).toMatchObject({
      startDay: '2026-12-20',
      endDay: '2026-12-26',
    });
    expect(withStart(dinner, '2026-12-06', '20:00')).toMatchObject({
      startDay: '2026-12-06',
      endDay: '2026-12-06',
      startTime: '20:00',
      endTime: '22:30',
    });
    // a late start runs past midnight
    expect(withStart(dinner, '2026-12-05', '23:00')).toMatchObject({
      endDay: '2026-12-06',
      endTime: '01:30',
    });
    expect(shifted(trip, 7)).toMatchObject({ startDay: '2026-12-19', endDay: '2026-12-25' });
  });

  it('switches between all day and a time', () => {
    expect(withAllDay(dinner, true)).toMatchObject({ allDay: true, startDay: '2026-12-05' });
    expect(withAllDay(trip, false)).toMatchObject({
      allDay: false,
      endDay: '2026-12-18',
      endTime: '18:00',
    });
    expect(withAllDay(whenOn('2026-12-05', true), false).endTime).toBe('');
  });

  it('says what is wrong', () => {
    expect(problem(dinner)).toBe('');
    expect(problem({ ...dinner, endTime: '17:00' })).toMatch(/end/);
    expect(problem({ ...trip, endDay: '2026-12-01' })).toMatch(/last day/);
  });

  it('says how long', () => {
    expect(lengthLabel(trip)).toBe('7 days · 6 nights');
    expect(lengthLabel(dinner)).toBe('2 h 30 min');
    expect(lengthLabel({ ...trip, endDay: trip.startDay })).toBe('');
  });

  it('compares dates by when they are, not how they are written', () => {
    const a = datesOf(dinner);
    expect(sameDates(a, { ...a, start: a.start.replace('.000Z', 'Z') })).toBe(true);
    expect(sameDates(a, datesOf(shifted(dinner, 1)))).toBe(false);
  });
});

describe('shiftBy', () => {
  it('moves an entry by days, keeping its times', () => {
    const e = datesOf(dinner);
    expect(shiftBy(e, 2)).toEqual({
      start: instantValue('2026-12-07T18:30'),
      end: instantValue('2026-12-07T21:00'),
    });
    expect(shiftBy(datesOf(trip), -2)).toEqual({
      start: allDayValue('2026-12-10'),
      end: allDayValue('2026-12-16'),
    });
  });
});
