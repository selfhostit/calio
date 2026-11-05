import { describe, expect, it } from 'vitest';
import { lookOf, offered, type Category } from './categories';

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
