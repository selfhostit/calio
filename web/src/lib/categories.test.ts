import { describe, expect, it } from 'vitest';
import { birthdayCategory, isBirthday, lookOf, offered, type Category } from './categories';

const cat = (over: Partial<Category>) =>
  ({
    id: 'c',
    name: 'Babysitting',
    color: 'rose',
    icon: 'baby',
    kind: 'petsit',
    enabled: true,
    sort: 0,
    created: '',
    ...over,
  }) as Category;

describe('lookOf', () => {
  it("uses the entry's category", () => {
    const shown = lookOf({ kind: 'petsit', category: 'c' }, [cat({})]);
    expect(shown.name).toBe('Babysitting');
    expect(shown.icon).toBe('baby');
    expect(shown.dot).toBe('bg-rose-400');
  });

  it('falls back on the kind when the category is gone', () => {
    const shown = lookOf({ kind: 'trip', category: 'gone' }, []);
    expect(shown.name).toBe('Trip');
    expect(shown.dot).toBe('bg-teal-400');
  });
});

describe('offered', () => {
  it('offers what is switched on, and the current one', () => {
    const on = cat({ id: 'on', sort: 2 });
    const off = cat({ id: 'off', enabled: false, sort: 1 });
    const first = cat({ id: 'first', sort: 0 });
    expect(offered([on, off, first]).map((c) => c.id)).toEqual(['first', 'on']);
    expect(offered([on, off, first], 'off').map((c) => c.id)).toEqual(['first', 'off', 'on']);
  });
});

describe('birthdays', () => {
  it('spots a birthday in a title', () => {
    for (const t of ["Maria's Birthday", 'Leo bday', 'Verjaardag oma', 'Γενέθλια Νίκου'])
      expect(isBirthday(t)).toBe(true);
    for (const t of ['Dentist', 'Bdaynamic meeting']) expect(isBirthday(t)).toBe(false);
  });
  it('finds the birthday category, if it is switched on', () => {
    const event = cat({ id: 'e', name: 'Event', icon: 'sparkle', kind: 'event' });
    const cake = cat({ id: 'b', name: 'Birthday', icon: 'cake', kind: 'event', sort: 4 });
    expect(birthdayCategory([event, cake])?.id).toBe('b');
    expect(birthdayCategory([event, { ...cake, name: 'Parties' }])?.id).toBe('b');
    expect(birthdayCategory([event, { ...cake, enabled: false }])).toBeUndefined();
  });
});
