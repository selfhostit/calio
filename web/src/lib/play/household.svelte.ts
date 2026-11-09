import { addDays } from '../days';
import { errorMessage, pb } from '../pb';
import { follow } from '../records';
import { session } from '../session.svelte';
import { today } from '../today.svelte';
import type { Companion } from './companion';
import { doneCount, isDue, type Check, type Habit } from './habits';
import { moodOf, moodScore, type Mood } from './mood';
import { wearing, type Owned, type Stage } from './shop';

// Everyone else's companions, for the household shelf: how they're doing,
// what they wear, and whether you've cheered them on today.

export interface Pal {
  companion: Companion;
  mood: Mood;
  stage: Stage;
  wearing: string[];
  done: number;
  due: number;
  cheered: boolean;
}

class Household {
  companions = $state<Companion[]>([]);
  habits = $state<Habit[]>([]);
  checks = $state<Check[]>([]);
  owned = $state<Owned[]>([]);
  /** The people you cheered on today. */
  cheered = $state<string[]>([]);
  error = $state('');

  #stop: (() => void) | null = null;

  get pals(): Pal[] {
    const day = today.day;
    return this.companions
      .map((companion) => {
        const habits = this.habits.filter((h) => h.user === companion.user && !h.archived);
        const due = habits.filter((h) => isDue(h, day, this.checks));
        return {
          companion,
          mood: moodOf(moodScore(habits, this.checks, day)),
          stage: companion.stage || 'cub',
          wearing: wearing(this.owned.filter((o) => o.user === companion.user)),
          done: due.filter((h) => doneCount(h, day, this.checks) >= h.times).length,
          due: due.length,
          cheered: this.cheered.includes(companion.user),
        };
      })
      .sort((a, b) => a.companion.name.localeCompare(b.companion.name));
  }

  async start() {
    const me = session.user?.id;
    if (!me || this.#stop) return;
    const others = pb.filter('user != {:me}', { me });
    const theirChecks = pb.filter('habit.user != {:me} && day >= {:from}', {
      me,
      from: addDays(today.day, -3),
    });
    try {
      const [companions, habits, checks, owned, cheers] = await Promise.all([
        pb.collection<Companion>('companions').getFullList({ filter: others }),
        pb.collection<Habit>('habits').getFullList({ filter: others }),
        pb.collection<Check>('habit_checks').getFullList({ filter: theirChecks, batch: 500 }),
        pb.collection<Owned>('inventory').getFullList({ filter: `${others} && equipped = true` }),
        pb.collection('cheers').getFullList({
          filter: pb.filter('from = {:me} && day = {:day}', { me, day: today.day }),
        }),
      ]);
      this.companions = companions;
      this.habits = habits;
      this.checks = checks;
      this.owned = owned;
      this.cheered = cheers.map((c) => c.to as string);
    } catch (err) {
      this.error = errorMessage(err, "Couldn't load the household's companions.");
      return;
    }
    const stops = await Promise.all([
      follow<Companion>('companions', (fn) => (this.companions = fn(this.companions)), others),
      follow<Habit>('habits', (fn) => (this.habits = fn(this.habits)), others),
      follow<Check>(
        'habit_checks',
        (fn) => (this.checks = fn(this.checks)),
        pb.filter('habit.user != {:me}', { me }),
      ),
      follow<Owned>('inventory', (fn) => (this.owned = fn(this.owned)), others),
    ]);
    this.#stop = () => stops.forEach((stop) => stop());
  }

  stop() {
    this.#stop?.();
    this.#stop = null;
  }

  /** Cheers someone on: they get a notification. Once a day. */
  async cheer(user: string) {
    await pb.send(`/api/app/play/cheer/${user}`, { method: 'POST' });
    this.cheered = [...this.cheered, user];
  }
}

export const household = new Household();
