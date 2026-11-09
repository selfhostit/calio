import { describe, expect, it } from 'vitest';
import { addDays, daysBetween, inDays, monthGrid } from './days';

describe('monthGrid', () => {
  it('shows six weeks starting on a Monday', () => {
    const days = monthGrid(2026, 11); // December 2026 starts on a Tuesday
    expect(days).toHaveLength(42);
    expect(days[0]).toBe('2026-11-30');
    expect(days[1]).toBe('2026-12-01');
  });
});

describe('addDays', () => {
  it('crosses months and years', () => {
    expect(addDays('2026-12-31', 1)).toBe('2027-01-01');
    expect(addDays('2026-03-01', -1)).toBe('2026-02-28');
  });
});

describe('daysBetween', () => {
  it('counts calendar days, whatever the clocks do', () => {
    // the last Sunday of March is 23 hours long in Europe
    expect(daysBetween('2026-03-28', '2026-03-30')).toBe(2);
    expect(daysBetween('2026-03-30', '2026-03-28')).toBe(-2);
  });
});

describe('inDays', () => {
  it('says how far away a day is', () => {
    expect(inDays('2026-11-09', '2026-11-09')).toBe('today');
    expect(inDays('2026-11-10', '2026-11-09')).toBe('tomorrow');
    expect(inDays('2026-11-12', '2026-11-09')).toBe('in 3 days');
  });
});
