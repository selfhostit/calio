import type { Kind } from './entries';

// How each kind of entry looks: soft tints with readable text, so the
// calendar stays calm when it's full (bar: the thin month bars on phones). Tailwind only sees whole class names, so
// they are spelled out here.
export const LOOK: Record<
  Kind,
  { icon: string; tile: string; pill: string; bar: string; chip: string; dot: string; text: string }
> = {
  event: {
    icon: 'sparkle',
    tile: 'bg-indigo-50 text-indigo-600 dark:bg-indigo-400/10 dark:text-indigo-300',
    pill: 'bg-indigo-100 text-indigo-900 dark:bg-indigo-400/20 dark:text-indigo-100',
    chip: 'border-indigo-300 bg-indigo-50 text-indigo-900 dark:border-indigo-400/50 dark:bg-indigo-400/15 dark:text-indigo-100',
    dot: 'bg-indigo-400',
    bar: 'max-sm:bg-indigo-300 dark:max-sm:bg-indigo-400/70',
    text: 'text-indigo-600 dark:text-indigo-300',
  },
  petsit: {
    icon: 'paw',
    tile: 'bg-amber-50 text-amber-700 dark:bg-amber-400/10 dark:text-amber-300',
    pill: 'bg-amber-100 text-amber-900 dark:bg-amber-400/20 dark:text-amber-100',
    chip: 'border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-400/50 dark:bg-amber-400/15 dark:text-amber-100',
    dot: 'bg-amber-400',
    bar: 'max-sm:bg-amber-300 dark:max-sm:bg-amber-400/70',
    text: 'text-amber-700 dark:text-amber-300',
  },
  trip: {
    icon: 'plane',
    tile: 'bg-teal-50 text-teal-700 dark:bg-teal-400/10 dark:text-teal-300',
    pill: 'bg-teal-100 text-teal-900 dark:bg-teal-400/20 dark:text-teal-100',
    chip: 'border-teal-300 bg-teal-50 text-teal-900 dark:border-teal-400/50 dark:bg-teal-400/15 dark:text-teal-100',
    dot: 'bg-teal-400',
    bar: 'max-sm:bg-teal-300 dark:max-sm:bg-teal-400/70',
    text: 'text-teal-700 dark:text-teal-300',
  },
};
