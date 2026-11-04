import type { Kind } from './entries';

// How each kind of entry looks. Tailwind only sees whole class names, so they
// are spelled out here.
export const LOOK: Record<
  Kind,
  { icon: string; tile: string; bar: string; pill: string; chip: string; dot: string }
> = {
  event: {
    icon: 'sparkle',
    tile: 'bg-violet-100 text-violet-700 dark:bg-violet-500/20 dark:text-violet-300',
    bar: 'bg-violet-500',
    pill: 'bg-violet-500 text-white',
    chip: 'border-violet-500 bg-violet-500 text-white',
    dot: 'bg-violet-500',
  },
  petsit: {
    icon: 'paw',
    tile: 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300',
    bar: 'bg-amber-500',
    pill: 'bg-amber-400 text-amber-950',
    chip: 'border-amber-500 bg-amber-500 text-white',
    dot: 'bg-amber-500',
  },
  trip: {
    icon: 'plane',
    tile: 'bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-300',
    bar: 'bg-sky-500',
    pill: 'bg-sky-500 text-white',
    chip: 'border-sky-500 bg-sky-500 text-white',
    dot: 'bg-sky-500',
  },
};
