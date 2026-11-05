import type { RecordModel } from 'pocketbase';
import { pb } from './pb';

// Notes and files about what the household looks after: "Wolfy" for the dog,
// "The house" for home sitting. A document can be shared with a sitter
// through a private link (/api/app/documents/{id}/share).

export interface Doc extends RecordModel {
  title: string;
  icon: string;
  body: string;
  files: string[];
  shared: boolean;
  created_by: string;
}

/** The address of one of a document's files. */
export function fileUrl(d: Doc, name: string, thumb = ''): string {
  return pb.files.getURL(d, name, thumb ? { thumb } : {});
}

const IMAGE = /\.(jpe?g|png|gif|webp|avif)$/i;
export const isImage = (name: string) => IMAGE.test(name);

/** A file's name without the random part PocketBase adds ("wolfy_ab12cd34ef.jpg" is "wolfy.jpg"). */
export function fileLabel(name: string): string {
  return name.replace(/_[a-zA-Z0-9]{10}(\.[^.]*)?$/, '$1');
}

/** A phone number as a tel: link, keeping only what dialers understand. */
export function telLink(phone: string): string {
  return 'tel:' + phone.replace(/[^\d+*#,;]/g, '');
}

export type Part = { text: string; href?: string };

const LINKS = /(https?:\/\/[^\s<>"]+[^\s<>".,;:!?)])|(\+?\d[\d ().-]{6,}\d)/g;

/** Splits text into plain parts and links: web addresses, and phone numbers to call. */
export function linkParts(text: string): Part[] {
  const parts: Part[] = [];
  let last = 0;
  for (const m of text.matchAll(LINKS)) {
    const at = m.index ?? 0;
    if (at > last) parts.push({ text: text.slice(last, at) });
    parts.push({ text: m[0], href: m[1] ? m[1] : telLink(m[0]) });
    last = at + m[0].length;
  }
  if (last < text.length) parts.push({ text: text.slice(last) });
  return parts;
}

/** Ideas for a new document, so an empty page isn't daunting. */
export const TEMPLATES: { title: string; icon: string; body: string }[] = [
  {
    title: 'Our pet',
    icon: 'paw',
    body: 'Food: \nWalks: \nMedicine: \nVet: \nHabits and quirks: \n',
  },
  {
    title: 'The house',
    icon: 'home',
    body: 'Keys and alarm: \nWi-Fi: \nPlants: \nBins go out on: \nIn an emergency: \n',
  },
  {
    title: 'The kids',
    icon: 'baby',
    body: 'Bedtime: \nFood and allergies: \nScreen time: \nEmergency contacts: \n',
  },
];
