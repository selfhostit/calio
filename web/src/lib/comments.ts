import type { RecordModel } from 'pocketbase';
import { dayOf, parseDate } from './entries';

// Comments on an entry; everyone else in the household is told about a new one.

export interface Comment extends RecordModel {
  entry: string;
  text: string;
  author: string;
}

const clock = new Intl.DateTimeFormat(undefined, { hour: 'numeric', minute: '2-digit' });
const date = new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short' });

/** When a comment was written: "just now", "5 min ago", "14:05", "Yesterday 18:30", "3 Dec". */
export function ago(created: string, now = new Date()): string {
  const d = parseDate(created);
  if (!d) return '';
  const minutes = Math.floor((now.getTime() - d.getTime()) / 60_000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes} min ago`;
  if (dayOf(d) === dayOf(now)) return clock.format(d);
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (dayOf(d) === dayOf(yesterday)) return `Yesterday ${clock.format(d)}`;
  return date.format(d);
}
