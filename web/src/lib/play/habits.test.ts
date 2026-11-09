import { describe, expect, it } from 'vitest';
import { isDue, progress, scheduleLabel, type Check, type Habit } from './habits';
import { moodOf, moodScore } from './mood';

let n = 0;
const habit = (fields: Partial<Habit>): Habit =>
  ({
    id: `h${n++}`,
    collectionId: '',
    collectionName: 'habits',
    user: 'u',
    title: 'x',
    icon: 'pill',
    times: 1,
    days: [],
    part: '',
    sort: 0,
    archived: false,
    created: '2026-11-01 08:00:00.000Z',
    ...fields,
  }) as Habit;

const check = (h: Habit, day: string, count: number, skipped = false): Check =>
  ({
    id: `${h.id}${day}`,
    collectionId: '',
    collectionName: 'habit_checks',
    habit: h.id,
    day,
    count,
    skipped,
  }) as Check;

describe('isDue', () => {
  it('follows the weekdays, and starts the day it was added', () => {
    const gym = habit({ days: ['mon', 'thu'] });
    expect(isDue(gym, '2026-11-09', [])).toBe(true); // a Monday
    expect(isDue(gym, '2026-11-10', [])).toBe(false);
    expect(isDue(gym, '2026-10-26', [])).toBe(false); // before it was added
  });

  it('is not due when skipped or put away', () => {
    const h = habit({});
    expect(isDue(h, '2026-11-09', [check(h, '2026-11-09', 0, true)])).toBe(false);
    expect(isDue(habit({ archived: true }), '2026-11-09', [])).toBe(false);
  });
});

describe('progress', () => {
  it('counts brushing once of twice as half', () => {
    const teeth = habit({ times: 2 });
    const d = habit({});
    expect(progress([teeth, d], [check(teeth, '2026-11-09', 1)], '2026-11-09')).toBe(0.25);
  });

  it('is null when nothing is due', () => {
    expect(progress([habit({ days: ['sun'] })], [], '2026-11-09')).toBeNull();
  });
});

describe('mood', () => {
  const h = habit({});
  const days = (...done: string[]) => done.map((d) => check(h, d, 1));

  it('starts happy for someone new', () => {
    expect(moodOf(moodScore([], [], '2026-11-09'))).toBe('happy');
  });

  it('gets sad after quiet days, and one good day cheers it up', () => {
    expect(moodOf(moodScore([h], [], '2026-11-09'))).toBe('sad');
    expect(moodOf(moodScore([h], days('2026-11-09'), '2026-11-09'))).toBe('happy');
  });

  it("doesn't start the morning sad after a good day", () => {
    expect(moodOf(moodScore([h], days('2026-11-07', '2026-11-08'), '2026-11-09'))).toBe('glowing');
    // one good day after quiet ones: fine, not sad
    expect(moodOf(moodScore([h], days('2026-11-08'), '2026-11-09'))).toBe('okay');
  });
});

describe('scheduleLabel', () => {
  it('says when', () => {
    expect(scheduleLabel({ days: [], times: 2 })).toBe('Every day · twice a day');
    expect(scheduleLabel({ days: ['mon', 'tue', 'wed', 'thu', 'fri'], times: 1 })).toBe('Weekdays');
    expect(scheduleLabel({ days: ['mon', 'thu'], times: 1 })).toBe('Mo, Th');
  });
});
