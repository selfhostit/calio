import { calendar } from './calendar.svelte';
import { addDays, dayLabel } from './days';
import { allDayValue, shiftBy, span, type Entry } from './entries';
import { errorMessage } from './pb';
import { toasts } from './toast.svelte';
import { sameDates, type Dates } from './when';

// Moving an entry to another day or time, from the form or by dragging it in
// the month, always ends with a note saying where it went and a way back.

type Moved = Dates & Pick<Entry, 'repeat_until'>;

const datesOf = (e: Entry): Moved => ({
  all_day: e.all_day,
  start: e.start,
  end: e.end,
  repeat_until: e.repeat_until,
});

/**
 * Moves an entry a number of days, keeping its times and length. A repeat
 * moves the whole series, and its last repeat moves along with it.
 */
export async function moveBy(e: Entry, days: number) {
  if (!days) return;
  const until = e.repeat_until
    ? { repeat_until: allDayValue(addDays(e.repeat_until.slice(0, 10), days)) }
    : {};
  try {
    const saved = await calendar.save(e.id, { ...shiftBy(e, days), ...until });
    offerUndo(e, saved);
  } catch (err) {
    toasts.show(errorMessage(err, `Couldn't move “${e.title}”.`));
  }
}

/** After an entry was saved: if it moved, says where to and offers to put it back. */
export function offerUndo(before: Entry, after: Entry) {
  if (sameDates(before, after)) return;
  const [first] = span(after);
  const where = dayLabel(first, calendar.today);
  const repeats = after.repeat ? ', and its repeats' : '';
  toasts.show(`Moved “${after.title}”${repeats} to ${where}`, {
    label: 'Undo',
    run: async () => {
      try {
        await calendar.save(after.id, datesOf(before));
        toasts.show(`“${after.title}” is back on ${dayLabel(span(before)[0], calendar.today)}`);
      } catch (err) {
        toasts.show(errorMessage(err, "Couldn't put it back."));
      }
    },
  });
}
