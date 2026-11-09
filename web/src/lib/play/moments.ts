import { addDays, type Day } from '../days';

// Moments: what's on the household's calendar that a companion reacts to.
// A trip coming up (it packs), being away (it sends a postcard), pet
// sitting (it has a visitor) and a birthday (party hat). The calendar hands
// these in as plain happenings (see lib/calendarMoments.ts), so this doesn't
// depend on how the calendar stores things.

export interface Happening {
  title: string;
  kind: 'trip' | 'petsit' | 'birthday' | 'other';
  first: Day;
  last: Day;
}

export type MomentKind = 'packing' | 'away' | 'visitor' | 'birthday';

export interface Moment {
  kind: MomentKind;
  line: string;
}

/** The one moment that matters most today, if any. */
export function momentOf(happenings: Happening[], today: Day): Moment | null {
  const on = (h: Happening) => h.first <= today && today <= h.last;
  const birthday = happenings.find((h) => h.kind === 'birthday' && on(h));
  if (birthday) return { kind: 'birthday', line: `Party hat on: ${birthday.title}!` };
  const away = happenings.find((h) => h.kind === 'trip' && on(h));
  if (away) return { kind: 'away', line: `Sending a postcard from ${away.title}!` };
  const visitor = happenings.find((h) => h.kind === 'petsit' && on(h));
  if (visitor) return { kind: 'visitor', line: 'We have a visitor today!' };
  const soon = happenings
    .filter((h) => h.kind === 'trip' && h.first > today && h.first <= addDays(today, 2))
    .sort((a, b) => (a.first < b.first ? -1 : 1))[0];
  if (soon)
    return {
      kind: 'packing',
      line: `Packing for ${soon.title}, ${soon.first === addDays(today, 1) ? 'tomorrow' : 'in two days'}!`,
    };
  return null;
}
