import { birthdayCategory, isBirthday, type Category } from './categories';
import { addDays, type Day } from './days';
import { occurrences, span, type Entry } from './entries';
import type { Happening } from './play/moments';

// What the calendar holds around today, as happenings the companion reacts
// to (lib/play/moments.ts).
export function happenings(entries: Entry[], categories: Category[], today: Day): Happening[] {
  const birthdays = birthdayCategory(categories)?.id;
  return occurrences(entries, today, addDays(today, 2)).map((e) => {
    const [first, last] = span(e);
    const kind =
      e.kind === 'trip'
        ? 'trip'
        : e.kind === 'petsit'
          ? 'petsit'
          : (birthdays && e.category === birthdays) || isBirthday(e.title)
            ? 'birthday'
            : 'other';
    return { title: e.title, kind, first, last };
  });
}
