import { addDays } from '../days';
import { errorMessage, pb } from '../pb';
import { follow, put, upsert } from '../records';
import { session } from '../session.svelte';
import { today } from '../today.svelte';
import type { Companion, Species } from './companion';
import { checkOf, type Check, type Habit } from './habits';
import { lineFor, moodOf, moodScore } from './mood';

// Your habits, the last week of checks and your companion, kept live. It
// loads when the Companion tab first opens, so the calendar stays as quick.

/** How far back checks are loaded: enough for the mood, and a week to show. */
const DAYS_BACK = 7;

class Play {
  habits = $state<Habit[]>([]);
  checks = $state<Check[]>([]);
  companion = $state<Companion | null>(null);
  loaded = $state(false);
  error = $state('');

  #stop: (() => void) | null = null;
  #user = '';
  /** Ticks save one after another, so quick taps don't race. */
  #saving: Promise<unknown> = Promise.resolve();

  /** Your habits that are still in use, in their order. */
  get active(): Habit[] {
    return this.habits
      .filter((h) => !h.archived)
      .sort((a, b) => a.sort - b.sort || (a.created < b.created ? -1 : 1));
  }

  get score(): number {
    return moodScore(this.active, this.checks, today.day);
  }

  get mood() {
    return moodOf(this.score);
  }

  line(allDone: boolean): string {
    return lineFor(this.mood, today.day, allDone);
  }

  async start() {
    const me = session.user?.id;
    if (!me || (this.#stop && this.#user === me)) return;
    this.stop();
    this.#user = me;
    this.error = '';
    const mine = pb.filter('user = {:me}', { me });
    const myChecks = pb.filter('habit.user = {:me}', { me });
    try {
      const [habits, checks, companions] = await Promise.all([
        pb.collection<Habit>('habits').getFullList({ filter: mine }),
        pb.collection<Check>('habit_checks').getFullList({
          filter: `${myChecks} && ${pb.filter('day >= {:from}', { from: addDays(today.day, -DAYS_BACK) })}`,
          batch: 500,
        }),
        pb.collection<Companion>('companions').getFullList({ filter: mine }),
      ]);
      this.habits = habits;
      this.checks = checks;
      this.companion = companions[0] ?? null;
      this.loaded = true;
    } catch (err) {
      this.error = errorMessage(err, "Couldn't load your habits.");
      return;
    }
    const stops = await Promise.all([
      follow<Habit>('habits', (fn) => (this.habits = fn(this.habits)), mine),
      follow<Check>('habit_checks', (fn) => (this.checks = fn(this.checks)), myChecks),
      follow<Companion>(
        'companions',
        (fn) => (this.companion = fn(this.companion ? [this.companion] : [])[0] ?? null),
        mine,
      ),
    ]);
    const unfollow = today.follow();
    this.#stop = () => {
      stops.forEach((stop) => stop());
      unfollow();
    };
  }

  stop() {
    this.#stop?.();
    this.#stop = null;
    this.loaded = false;
  }

  async meet(name: string, species: Species, starters: Partial<Habit>[]) {
    const user = this.#user;
    this.companion = await pb
      .collection<Companion>('companions')
      .create({ user, name: name.trim(), species });
    for (const [i, h] of starters.entries()) await this.saveHabit(null, { ...h, sort: i });
  }

  async rename(name: string) {
    if (!this.companion) return;
    this.companion = await pb
      .collection<Companion>('companions')
      .update(this.companion.id, { name: name.trim() });
  }

  async saveHabit(id: string | null, data: Partial<Habit>): Promise<Habit> {
    const fields = id ? data : { user: this.#user, sort: this.habits.length, ...data };
    const saved = await upsert(pb.collection<Habit>('habits'), id, fields);
    this.habits = put(this.habits, saved);
    return saved;
  }

  /** Puts a habit away; its history stays, and it no longer counts. */
  async archive(h: Habit) {
    await this.saveHabit(h.id, { archived: true });
  }

  /**
   * Ticks a habit off once more today; ticking one that's done starts it
   * over. Shows right away, and goes back if saving fails.
   */
  tick(h: Habit): Promise<void> {
    return this.#queue(() => {
      const c = checkOf(this.checks, h, today.day);
      const count = (c?.count ?? 0) >= h.times ? 0 : (c?.count ?? 0) + 1;
      return this.#setCheck(h, { count, skipped: false }, c);
    });
  }

  /** Skips a habit today, or takes the skip back. */
  skip(h: Habit, skipped: boolean): Promise<void> {
    return this.#queue(() => this.#setCheck(h, { skipped }, checkOf(this.checks, h, today.day)));
  }

  #queue(fn: () => Promise<void>): Promise<void> {
    const next = this.#saving.then(fn);
    this.#saving = next.catch(() => {});
    return next;
  }

  async #setCheck(h: Habit, data: Partial<Check>, c: Check | undefined) {
    const before = this.checks;
    const day = today.day;
    if (c) this.checks = put(this.checks, { ...c, ...data });
    else
      this.checks = [
        ...this.checks,
        { id: `new-${h.id}`, habit: h.id, day, count: 0, skipped: false, ...data } as Check,
      ];
    try {
      const saved = c
        ? await pb.collection<Check>('habit_checks').update(c.id, data)
        : await pb.collection<Check>('habit_checks').create({ habit: h.id, day, ...data });
      this.checks = put(
        this.checks.filter((x) => x.id !== `new-${h.id}`),
        saved,
      );
    } catch (err) {
      this.checks = before;
      throw err;
    }
  }
}

export const play = new Play();
