import { describe, expect, it } from 'vitest';
import { ago } from './comments';

describe('ago', () => {
  const now = new Date(2026, 11, 12, 18, 0);
  const at = (d: Date) => d.toISOString().replace('T', ' ');

  it('says how long ago, then when', () => {
    expect(ago(at(new Date(2026, 11, 12, 17, 59, 40)), now)).toBe('just now');
    expect(ago(at(new Date(2026, 11, 12, 17, 55)), now)).toBe('5 min ago');
    expect(ago(at(new Date(2026, 11, 11, 9, 0)), now)).toMatch(/^Yesterday /);
    expect(ago(at(new Date(2026, 11, 3, 9, 0)), now)).toMatch(/3/);
  });
});
