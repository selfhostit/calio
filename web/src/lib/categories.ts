import type { RecordModel } from 'pocketbase';
import type { Entry, Kind } from './entries';
import { LOOKS, type Color, type Look } from './look';

// What the household puts on its calendar: events, pet sitting, trips,
// babysitting or anything it adds. A category's kind is how its entries
// behave: "petsit" ones need a sitter, "trip" ones get a packing list.

export interface Category extends RecordModel {
  name: string;
  color: Color;
  icon: string;
  kind: Kind;
  enabled: boolean;
  sort: number;
  created_by: string;
}

export const BEHAVIOURS: { kind: Kind; label: string; hint: string }[] = [
  { kind: 'event', label: 'Plain', hint: 'A dinner, an appointment, a birthday.' },
  { kind: 'petsit', label: 'Needs a sitter', hint: 'Tracks who is looking after them.' },
  { kind: 'trip', label: 'Trip', hint: 'Comes with a packing list.' },
];

/** How an entry looks when its category is gone: as its kind did before categories. */
const BY_KIND: Record<Kind, { name: string; color: Color; icon: string }> = {
  event: { name: 'Event', color: 'indigo', icon: 'sparkle' },
  petsit: { name: 'Pet sitting', color: 'amber', icon: 'paw' },
  trip: { name: 'Trip', color: 'teal', icon: 'plane' },
};

export type Shown = Look & { name: string; icon: string };

/** How an entry shows: its category's name, color and icon. */
export function lookOf(e: Pick<Entry, 'kind' | 'category'>, categories: Category[]): Shown {
  const c = categories.find((c) => c.id === e.category) ?? BY_KIND[e.kind] ?? BY_KIND.event;
  return { ...(LOOKS[c.color] ?? LOOKS.indigo), name: c.name, icon: c.icon };
}

export function bySort(a: Category, b: Category): number {
  return a.sort - b.sort || a.created.localeCompare(b.created);
}

/**
 * The categories to pick from for an entry: the ones switched on, and the
 * entry's own even when it's switched off.
 */
export function offered(categories: Category[], current = ''): Category[] {
  return categories.filter((c) => c.enabled || c.id === current).sort(bySort);
}

/** Whether a title is about a birthday (as backend/app/birthdays.go reads it). */
export function isBirthday(title: string): boolean {
  return /birth ?days?|\bb-?days?\b|verjaardag|geburtstag|cumpleaños|anniversaire|γενέθλια/i.test(
    title,
  );
}

/**
 * The switched-on category birthdays go in: one named Birthday, else one with
 * the cake icon.
 */
export function birthdayCategory(categories: Category[]): Category | undefined {
  const plain = offered(categories).filter((c) => c.kind === 'event');
  return plain.find((c) => isBirthday(c.name)) ?? plain.find((c) => c.icon === 'cake');
}
