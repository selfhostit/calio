import type { RecordModel } from 'pocketbase';
import { pb } from './pb';

// Bringing other calendars in: a .ics file, copied once, or a calendar's link,
// which the server checks every half hour (backend/app/imports.go).

/** A calendar the household subscribes to by its link. */
export interface Subscription extends RecordModel {
  name: string;
  category: string;
  /** When the server last fetched it, and what went wrong if it failed. */
  synced: string;
  error: string;
  /** How many entries it has on the calendar. */
  events: number;
  created_by: string;
}

export interface ImportResult {
  added: number;
  updated: number;
  total: number;
}

export function importFile(file: File, category: string): Promise<ImportResult> {
  const body = new FormData();
  body.append('file', file);
  if (category) body.append('category', category);
  return pb.send('/api/app/import', { method: 'POST', body });
}

export function subscribe(url: string, name: string, category: string): Promise<Subscription> {
  return pb.send('/api/app/calendars', { method: 'POST', body: { url, name, category } });
}

export function refresh(id: string): Promise<Subscription> {
  return pb.send(`/api/app/calendars/${id}/refresh`, { method: 'POST' });
}

/** "Added 12 events", "Nothing new: 12 events were already here". */
export function importSummary(r: ImportResult): string {
  const n = (k: number) => `${k} ${k === 1 ? 'event' : 'events'}`;
  if (r.total === 0) return 'That file has no events from the last year or later.';
  if (r.added === 0 && r.updated === 0) return `Nothing new: ${n(r.total)} were already here.`;
  const parts = [];
  if (r.added) parts.push(`added ${n(r.added)}`);
  if (r.updated) parts.push(`updated ${n(r.updated)}`);
  const s = parts.join(' and ');
  return s.charAt(0).toUpperCase() + s.slice(1) + '.';
}

/** Whether pasted text looks like a calendar link (https://, http:// or webcal://). */
export function looksLikeLink(s: string): boolean {
  return /^(https?|webcals?):\/\/[^\s/]+\.[^\s]+$/i.test(s.trim());
}

/** How to get a calendar out of another app: as a link, as a file, or both. */
export interface Guide {
  app: string;
  link?: string[];
  file?: string[];
  note?: string;
}

export const GUIDES: Guide[] = [
  {
    app: 'Google Calendar',
    link: [
      'On a computer, open calendar.google.com (the phone app can’t do this).',
      'Click the gear, then Settings.',
      'Under “Settings for my calendars” on the left, click the calendar.',
      'Scroll to “Integrate calendar” and copy “Secret address in iCal format”.',
    ],
    file: [
      'On a computer, open calendar.google.com, click the gear, then Settings.',
      'Click “Import & export”, then Export. Google downloads a .zip file.',
      'Unzip it: there is one .ics file for each of your calendars.',
    ],
    note: 'Anyone with the secret address can see that calendar. If it leaks, “Reset” in the same place makes a new one.',
  },
  {
    app: 'Outlook and Microsoft 365',
    link: [
      'Open Outlook on the web (outlook.com or outlook.office.com) and click the gear.',
      'Go to Calendar, then Shared calendars.',
      'Under “Publish a calendar”, pick the calendar and “Can view all details”, then Publish.',
      'Copy the ICS link.',
    ],
    file: [
      'In Outlook for Windows, open the calendar.',
      'Choose File, then Save Calendar, and pick the dates and “Full details”.',
    ],
    note: 'At work, your organization may have switched publishing off. Ask whoever runs your Microsoft 365.',
  },
  {
    app: 'Apple Calendar (iCloud)',
    link: [
      'On an iPhone or iPad, open Calendar and tap Calendars at the bottom.',
      'Tap the ⓘ next to the calendar and turn on Public Calendar.',
      'Tap Share Link, then Copy.',
      'On a Mac: in Calendar, Control-click the calendar, choose Sharing Settings, tick Public Calendar and copy the link.',
    ],
    file: [
      'On a Mac, open Calendar and click the calendar in the list.',
      'Choose File, then Export, then Export…, and save the .ics file.',
    ],
    note: 'A public iCloud calendar can be read by anyone who has the link.',
  },
  {
    app: 'Proton Calendar',
    link: [
      'Open Proton Calendar on the web, then Settings, then Calendars.',
      'Pick the calendar, then Share, and create a link with full details.',
      'Copy the link.',
    ],
    file: ['In Settings, then Calendars, pick the calendar and choose Export.'],
  },
  {
    app: 'Other apps',
    link: [
      'Fastmail, Nextcloud, Thunderbird, Zoho, school and team calendars: look for a link called iCal, ICS, webcal or “subscribe”.',
    ],
    file: ['Most calendar apps can export or save a calendar as an .ics file.'],
  },
];

/** When a subscription was last fetched: "just now", "5 min ago", "3 h ago", "2 Dec". */
export function syncedLabel(synced: string, now = new Date()): string {
  if (!synced) return 'not fetched yet';
  const t = new Date(synced.replace(' ', 'T'));
  const min = Math.round((now.getTime() - t.getTime()) / 60_000);
  if (min < 1) return 'just now';
  if (min < 60) return `${min} min ago`;
  if (min < 24 * 60) return `${Math.round(min / 60)} h ago`;
  return t.toLocaleDateString(undefined, { day: 'numeric', month: 'short' });
}
