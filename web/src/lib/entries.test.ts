import { describe, expect, it } from 'vitest';
import { addDays, parseDate } from './days';
import {
  agenda,
  allDayValue,
  mapLink,
  occurrences,
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
    repeat: '',
    repeat_until: '',
    checklist: null,
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

describe('progressLabel', () => {
  it('counts the days of a trip', () => {
    expect(progressLabel(trip, '2026-12-14')).toBe('Day 3 of 7');
    expect(progressLabel(trip, addDays('2026-12-18', 1))).toBe('');
  });
});

describe('occurrences', () => {
  const rent = entry({
    title: 'Rent',
    all_day: true,
    start: allDayValue('2026-01-31'),
    repeat: 'monthly',
  });

  it('skips months without the day', () => {
    const days = occurrences([rent], '2026-01-01', '2026-05-31').map((e) => span(e)[0]);
    expect(days).toEqual(['2026-01-31', '2026-03-31', '2026-05-31']);
  });

  it('stops at the last day', () => {
    const gym = entry({
      all_day: true,
      start: allDayValue('2026-01-01'),
      repeat: 'daily',
      repeat_until: allDayValue('2026-01-03'),
    });
    expect(occurrences([gym], '2026-01-01', '2026-02-01')).toHaveLength(3);
  });

  it('keeps the time of day and gives each repeat its own key', () => {
    const piano = entry({ start: new Date(2026, 0, 7, 17, 30).toISOString(), repeat: 'weekly' });
    const [, second] = occurrences([piano], '2026-01-01', '2026-01-20');
    expect(span(second)[0]).toBe('2026-01-14');
    expect(parseDate(second.start)!.getHours()).toBe(17);
    expect(second.key).not.toBe(piano.id);
  });

  it('keeps a multi-day repeat that started before the range', () => {
    const camp = entry({
      all_day: true,
      start: allDayValue('2025-07-30'),
      end: allDayValue('2025-08-03'),
      repeat: 'yearly',
    });
    expect(occurrences([camp], '2026-08-01', '2026-08-31').map((e) => span(e))).toEqual([
      ['2026-07-30', '2026-08-03'],
    ]);
  });
});

describe('mapLink', () => {
  it('searches for an address or coordinates', () => {
    expect(mapLink('37.98, 23.73')).toBe(
      'https://www.google.com/maps/search/?api=1&query=37.98%2C%2023.73',
    );
  });

  it('keeps a pasted link', () => {
    expect(mapLink('https://maps.app.goo.gl/abc')).toBe('https://maps.app.goo.gl/abc');
  });
});
