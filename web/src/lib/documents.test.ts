import { describe, expect, it } from 'vitest';
import { fileLabel, linkParts, telLink } from './documents';

describe('telLink', () => {
  it('keeps digits and a leading plus', () => {
    expect(telLink('+30 (210) 123-4567')).toBe('tel:+302101234567');
  });
});

describe('linkParts', () => {
  it('finds web addresses and phone numbers', () => {
    expect(linkParts('Vet: +32 470 12 34 56, or https://vet.example/wolfy.')).toEqual([
      { text: 'Vet: ' },
      { text: '+32 470 12 34 56', href: 'tel:+32470123456' },
      { text: ', or ' },
      { text: 'https://vet.example/wolfy', href: 'https://vet.example/wolfy' },
      { text: '.' },
    ]);
  });

  it('leaves short numbers alone', () => {
    expect(linkParts('2 cups at 18:00')).toEqual([{ text: '2 cups at 18:00' }]);
  });
});

describe('fileLabel', () => {
  it('drops the random part', () => {
    expect(fileLabel('wolfy_ab12cd34ef.jpg')).toBe('wolfy.jpg');
    expect(fileLabel('notes.pdf')).toBe('notes.pdf');
  });
});
