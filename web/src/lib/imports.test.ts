import { describe, expect, it } from 'vitest';
import { GUIDES, importSummary, looksLikeLink, syncedLabel } from './imports';

describe('looksLikeLink', () => {
  it('takes the links calendar apps hand out', () => {
    expect(
      looksLikeLink(
        'https://calendar.google.com/calendar/ical/x%40gmail.com/private-abc/basic.ics',
      ),
    ).toBe(true);
    expect(looksLikeLink('webcal://p52-caldav.icloud.com/published/2/abc')).toBe(true);
    expect(
      looksLikeLink(' https://outlook.office365.com/owa/calendar/abc/reachcalendar.ics '),
    ).toBe(true);
  });
  it('turns down what isn’t one', () => {
    expect(looksLikeLink('calendar.google.com')).toBe(false);
    expect(looksLikeLink('ftp://example.com/cal.ics')).toBe(false);
    expect(looksLikeLink('https://exa mple.com')).toBe(false);
  });
});

describe('importSummary', () => {
  it('says what changed', () => {
    expect(importSummary({ added: 12, updated: 0, total: 12 })).toBe('Added 12 events.');
    expect(importSummary({ added: 1, updated: 2, total: 3 })).toBe(
      'Added 1 event and updated 2 events.',
    );
    expect(importSummary({ added: 0, updated: 0, total: 4 })).toBe(
      'Nothing new: 4 events were already here.',
    );
    expect(importSummary({ added: 0, updated: 0, total: 0 })).toMatch(/no events/);
  });
});

describe('syncedLabel', () => {
  const now = new Date('2026-12-02T12:00:00Z');
  it('says how long ago', () => {
    expect(syncedLabel('', now)).toBe('not fetched yet');
    expect(syncedLabel('2026-12-02 11:59:40.000Z', now)).toBe('just now');
    expect(syncedLabel('2026-12-02 11:30:00.000Z', now)).toBe('30 min ago');
    expect(syncedLabel('2026-12-02 09:00:00.000Z', now)).toBe('3 h ago');
  });
});

describe('GUIDES', () => {
  it('cover the apps people use most, each with a way in', () => {
    const apps = GUIDES.map((g) => g.app).join(' ');
    for (const app of ['Google', 'Outlook', 'Apple']) expect(apps).toContain(app);
    for (const g of GUIDES)
      expect((g.link?.length ?? 0) + (g.file?.length ?? 0)).toBeGreaterThan(0);
  });
});
