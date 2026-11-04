import { describe, expect, it } from 'vitest';
import { findByName, telLink, type Contact } from './contacts';

describe('telLink', () => {
  it('keeps digits and a leading plus', () => {
    expect(telLink('+30 (210) 123-4567')).toBe('tel:+302101234567');
  });
});

describe('findByName', () => {
  const nikos = { name: 'Nikos ', role: 'petsitter' } as Contact;
  it('matches names whatever the case and spacing', () => {
    expect(findByName([nikos], ' nikos')).toBe(nikos);
    expect(findByName([nikos], '')).toBeUndefined();
  });
});
