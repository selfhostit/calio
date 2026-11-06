import { describe, expect, it } from 'vitest';
import { firstName } from './calendar.svelte';
import type { User } from './pb';

const person = (fields: Partial<User>) =>
  ({ id: 'u1', name: '', role: 'member', ...fields }) as User;

describe('firstName', () => {
  it('takes the first word of the name', () => {
    expect(firstName(person({ name: 'Ada Lovelace' }))).toBe('Ada');
  });
  it('falls back to your own email', () => {
    expect(firstName(person({ email: 'ada@example.com' }))).toBe('ada');
  });
  // PocketBase leaves out other people's email, so this used to throw and
  // stopped the entry form opening
  it('copes with someone who has no name and no email', () => {
    expect(firstName(person({}))).toBe('Someone');
  });
});
