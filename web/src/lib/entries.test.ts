import { describe, expect, it } from 'vitest';
import {
  addDays,
  agenda,
  allDayValue,
  monthGrid,
  progressLabel,
  span,
  type Entry,
} from './entries';

const entry = (fields: Partial<Entry>): Entry =>
  ({
    id: fields.title ?? 'x',
    collectionId: '',
    collectionName: 'entries',
    title: 'x',
    kind: 'event',
    start: '',
    end: '',
    all_day: false,
    location: '',
    notes: '',
    people: [],
    sitter: '',
    sitter_status: '',
    created_by: '',
    ...fields,
  }) as Entry;

const trip = entry({
  title: 'Crete',
  kind: 'trip',
  all_day: true,
  start: allDayValue('2026-12-12'),
  end: allDayValue('2026-12-18'),
});

describe('span', () => {
  it('reads all-day dates as they are', () => {
    expect(span(trip)).toEqual(['2026-12-12', '2026-12-18']);
  });

  it('uses the start when there is no end', () => {
    expect(span(entry({ all_day: true, start: allDayValue('2026-12-12') }))).toEqual([
      '2026-12-12',
      '2026-12-12',
    ]);
  });

  it('keeps an evening that ends at midnight on one day', () => {
    const start = new Date(2026, 11, 12, 20, 0).toISOString();
    const end = new Date(2026, 11, 13, 0, 0).toISOString();
    expect(span(entry({ start, end }))).toEqual(['2026-12-12', '2026-12-12']);
  });
});

describe('agenda', () => {
  const party = entry({ title: 'Party', start: new Date(2026, 11, 14, 19).toISOString() });
  const past = entry({ title: 'Past', all_day: true, start: allDayValue('2026-12-01') });

  it('drops what is over and puts what is going on under today', () => {
    expect(agenda([party, past, trip], '2026-12-13')).toEqual([
      { day: '2026-12-13', entries: [trip] },
      { day: '2026-12-14', entries: [party] },
    ]);
  });
});

describe('monthGrid', () => {
  it('shows six weeks starting on a Monday', () => {
    const days = monthGrid(2026, 11); // December 2026 starts on a Tuesday
    expect(days).toHaveLength(42);
    expect(days[0]).toBe('2026-11-30');
    expect(days[1]).toBe('2026-12-01');
  });
});

describe('progressLabel', () => {
  it('counts the days of a trip', () => {
    expect(progressLabel(trip, '2026-12-14')).toBe('Day 3 of 7');
    expect(progressLabel(trip, addDays('2026-12-18', 1))).toBe('');
  });
});
