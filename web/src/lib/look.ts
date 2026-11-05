// How a category looks: soft tints with readable text, so the calendar stays
// calm when it's full (bar: the thin month bars on phones). Tailwind only sees
// whole class names, so they are spelled out here for every color a category
// can have (the same list as CategoryColors in backend/migrations).

export const COLORS = [
  'indigo',
  'amber',
  'teal',
  'rose',
  'sky',
  'emerald',
  'violet',
  'orange',
  'lime',
  'pink',
  'cyan',
  'slate',
] as const;
export type Color = (typeof COLORS)[number];

export interface Look {
  tile: string;
  pill: string;
  chip: string;
  dot: string;
  bar: string;
  text: string;
}

export const LOOKS: Record<Color, Look> = {
  indigo: {
    tile: 'bg-indigo-50 text-indigo-600 dark:bg-indigo-400/10 dark:text-indigo-300',
    pill: 'bg-indigo-100 text-indigo-900 dark:bg-indigo-400/20 dark:text-indigo-100',
    chip: 'border-indigo-300 bg-indigo-50 text-indigo-900 dark:border-indigo-400/50 dark:bg-indigo-400/15 dark:text-indigo-100',
    dot: 'bg-indigo-400',
    bar: 'max-sm:bg-indigo-300 dark:max-sm:bg-indigo-400/70',
    text: 'text-indigo-600 dark:text-indigo-300',
  },
  amber: {
    tile: 'bg-amber-50 text-amber-700 dark:bg-amber-400/10 dark:text-amber-300',
    pill: 'bg-amber-100 text-amber-900 dark:bg-amber-400/20 dark:text-amber-100',
    chip: 'border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-400/50 dark:bg-amber-400/15 dark:text-amber-100',
    dot: 'bg-amber-400',
    bar: 'max-sm:bg-amber-300 dark:max-sm:bg-amber-400/70',
    text: 'text-amber-700 dark:text-amber-300',
  },
  teal: {
    tile: 'bg-teal-50 text-teal-700 dark:bg-teal-400/10 dark:text-teal-300',
    pill: 'bg-teal-100 text-teal-900 dark:bg-teal-400/20 dark:text-teal-100',
    chip: 'border-teal-300 bg-teal-50 text-teal-900 dark:border-teal-400/50 dark:bg-teal-400/15 dark:text-teal-100',
    dot: 'bg-teal-400',
    bar: 'max-sm:bg-teal-300 dark:max-sm:bg-teal-400/70',
    text: 'text-teal-700 dark:text-teal-300',
  },
  rose: {
    tile: 'bg-rose-50 text-rose-600 dark:bg-rose-400/10 dark:text-rose-300',
    pill: 'bg-rose-100 text-rose-900 dark:bg-rose-400/20 dark:text-rose-100',
    chip: 'border-rose-300 bg-rose-50 text-rose-900 dark:border-rose-400/50 dark:bg-rose-400/15 dark:text-rose-100',
    dot: 'bg-rose-400',
    bar: 'max-sm:bg-rose-300 dark:max-sm:bg-rose-400/70',
    text: 'text-rose-600 dark:text-rose-300',
  },
  sky: {
    tile: 'bg-sky-50 text-sky-600 dark:bg-sky-400/10 dark:text-sky-300',
    pill: 'bg-sky-100 text-sky-900 dark:bg-sky-400/20 dark:text-sky-100',
    chip: 'border-sky-300 bg-sky-50 text-sky-900 dark:border-sky-400/50 dark:bg-sky-400/15 dark:text-sky-100',
    dot: 'bg-sky-400',
    bar: 'max-sm:bg-sky-300 dark:max-sm:bg-sky-400/70',
    text: 'text-sky-600 dark:text-sky-300',
  },
  emerald: {
    tile: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-300',
    pill: 'bg-emerald-100 text-emerald-900 dark:bg-emerald-400/20 dark:text-emerald-100',
    chip: 'border-emerald-300 bg-emerald-50 text-emerald-900 dark:border-emerald-400/50 dark:bg-emerald-400/15 dark:text-emerald-100',
    dot: 'bg-emerald-400',
    bar: 'max-sm:bg-emerald-300 dark:max-sm:bg-emerald-400/70',
    text: 'text-emerald-700 dark:text-emerald-300',
  },
  violet: {
    tile: 'bg-violet-50 text-violet-600 dark:bg-violet-400/10 dark:text-violet-300',
    pill: 'bg-violet-100 text-violet-900 dark:bg-violet-400/20 dark:text-violet-100',
    chip: 'border-violet-300 bg-violet-50 text-violet-900 dark:border-violet-400/50 dark:bg-violet-400/15 dark:text-violet-100',
    dot: 'bg-violet-400',
    bar: 'max-sm:bg-violet-300 dark:max-sm:bg-violet-400/70',
    text: 'text-violet-600 dark:text-violet-300',
  },
  orange: {
    tile: 'bg-orange-50 text-orange-700 dark:bg-orange-400/10 dark:text-orange-300',
    pill: 'bg-orange-100 text-orange-900 dark:bg-orange-400/20 dark:text-orange-100',
    chip: 'border-orange-300 bg-orange-50 text-orange-900 dark:border-orange-400/50 dark:bg-orange-400/15 dark:text-orange-100',
    dot: 'bg-orange-400',
    bar: 'max-sm:bg-orange-300 dark:max-sm:bg-orange-400/70',
    text: 'text-orange-700 dark:text-orange-300',
  },
  lime: {
    tile: 'bg-lime-50 text-lime-700 dark:bg-lime-400/10 dark:text-lime-300',
    pill: 'bg-lime-100 text-lime-900 dark:bg-lime-400/20 dark:text-lime-100',
    chip: 'border-lime-300 bg-lime-50 text-lime-900 dark:border-lime-400/50 dark:bg-lime-400/15 dark:text-lime-100',
    dot: 'bg-lime-400',
    bar: 'max-sm:bg-lime-300 dark:max-sm:bg-lime-400/70',
    text: 'text-lime-700 dark:text-lime-300',
  },
  pink: {
    tile: 'bg-pink-50 text-pink-600 dark:bg-pink-400/10 dark:text-pink-300',
    pill: 'bg-pink-100 text-pink-900 dark:bg-pink-400/20 dark:text-pink-100',
    chip: 'border-pink-300 bg-pink-50 text-pink-900 dark:border-pink-400/50 dark:bg-pink-400/15 dark:text-pink-100',
    dot: 'bg-pink-400',
    bar: 'max-sm:bg-pink-300 dark:max-sm:bg-pink-400/70',
    text: 'text-pink-600 dark:text-pink-300',
  },
  cyan: {
    tile: 'bg-cyan-50 text-cyan-700 dark:bg-cyan-400/10 dark:text-cyan-300',
    pill: 'bg-cyan-100 text-cyan-900 dark:bg-cyan-400/20 dark:text-cyan-100',
    chip: 'border-cyan-300 bg-cyan-50 text-cyan-900 dark:border-cyan-400/50 dark:bg-cyan-400/15 dark:text-cyan-100',
    dot: 'bg-cyan-400',
    bar: 'max-sm:bg-cyan-300 dark:max-sm:bg-cyan-400/70',
    text: 'text-cyan-700 dark:text-cyan-300',
  },
  slate: {
    tile: 'bg-slate-50 text-slate-600 dark:bg-slate-400/10 dark:text-slate-300',
    pill: 'bg-slate-100 text-slate-900 dark:bg-slate-400/20 dark:text-slate-100',
    chip: 'border-slate-300 bg-slate-50 text-slate-900 dark:border-slate-400/50 dark:bg-slate-400/15 dark:text-slate-100',
    dot: 'bg-slate-400',
    bar: 'max-sm:bg-slate-300 dark:max-sm:bg-slate-400/70',
    text: 'text-slate-600 dark:text-slate-300',
  },
};

/** The icons a category or a document can have (see Icon.svelte). */
export const ICONS = [
  'sparkle',
  'paw',
  'plane',
  'baby',
  'home',
  'heart',
  'cake',
  'cart',
  'car',
  'work',
  'health',
  'school',
  'ball',
  'music',
  'food',
  'plant',
  'star',
  'user',
] as const;
