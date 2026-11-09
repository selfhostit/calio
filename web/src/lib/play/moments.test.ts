import { describe, expect, it } from 'vitest';
import { momentOf, type Happening } from './moments';

const h = (kind: Happening['kind'], first: string, last = first, title = 'x'): Happening => ({
  title,
  kind,
  first,
  last,
});

describe('momentOf', () => {
  it('packs before a trip and sends a postcard during it', () => {
    const crete = h('trip', '2026-11-11', '2026-11-15', 'Crete');
    expect(momentOf([crete], '2026-11-10')?.line).toBe('Packing for Crete, tomorrow!');
    expect(momentOf([crete], '2026-11-08')).toBeNull();
    expect(momentOf([crete], '2026-11-12')?.kind).toBe('away');
  });

  it('puts a birthday first', () => {
    expect(
      momentOf(
        [h('petsit', '2026-11-09'), h('birthday', '2026-11-09', '2026-11-09', "Anna's birthday")],
        '2026-11-09',
      ),
    ).toEqual({ kind: 'birthday', line: "Party hat on: Anna's birthday!" });
  });
});
