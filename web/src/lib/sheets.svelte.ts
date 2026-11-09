import { calendar } from './calendar.svelte';
import type { Day } from './days';
import type { Entry } from './entries';

// The sheets that open over whatever section is showing: an entry (opened
// from the calendar, or from today's summary at the top) and settings.
class Sheets {
  entry = $state<{ entry: Entry | null; day: Day } | null>(null);
  /** Open, and whether to go straight to importing calendars. */
  settings = $state<false | 'open' | 'import'>(false);

  /** Opens an entry; a repeat opens the entry it repeats, so edits apply to all of them. */
  open(entry: Entry) {
    this.entry = {
      entry: calendar.entries.find((e) => e.id === entry.id) ?? entry,
      day: calendar.today,
    };
  }

  /** Opens the form for a new entry on a day. */
  add(day: Day = calendar.today) {
    this.entry = { entry: null, day };
  }
}

export const sheets = new Sheets();
