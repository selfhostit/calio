import { dayOf, type Entry } from './entries';
import { errorMessage, pb, type User } from './pb';

// The household's calendar and its people, kept live: a change anyone makes
// arrives here through PocketBase's realtime subscription.
class Calendar {
  entries = $state<Entry[]>([]);
  people = $state<User[]>([]);
  loaded = $state(false);
  error = $state('');
  today = $state(dayOf(new Date()));

  #stop: (() => void) | null = null;

  async start() {
    const entries = pb.collection<Entry>('entries');
    try {
      const [list, people] = await Promise.all([
        entries.getFullList({ sort: 'start', batch: 500 }),
        pb.collection<User>('users').getFullList({ sort: 'name' }),
      ]);
      this.entries = list;
      this.people = people;
      this.loaded = true;
    } catch (err) {
      this.error = errorMessage(err, "Couldn't load the calendar.");
      return;
    }

    const unsubscribe = await entries.subscribe('*', ({ action, record }) => {
      const rest = this.entries.filter((e) => e.id !== record.id);
      this.entries = action === 'delete' ? rest : [...rest, record];
    });
    // the day changes at midnight, and when the phone wakes the app up
    const tick = () => (this.today = dayOf(new Date()));
    const timer = setInterval(tick, 60_000);
    document.addEventListener('visibilitychange', tick);
    this.#stop = () => {
      unsubscribe();
      clearInterval(timer);
      document.removeEventListener('visibilitychange', tick);
    };
  }

  stop() {
    this.#stop?.();
    this.#stop = null;
  }

  person(id: string): User | undefined {
    return this.people.find((p) => p.id === id);
  }

  async save(id: string | null, data: Partial<Entry>): Promise<Entry> {
    const entries = pb.collection<Entry>('entries');
    const saved = id ? await entries.update(id, data) : await entries.create(data);
    // show it right away; the realtime event that follows is the same record
    this.entries = [...this.entries.filter((e) => e.id !== saved.id), saved];
    return saved;
  }

  async remove(id: string) {
    await pb.collection('entries').delete(id);
    this.entries = this.entries.filter((e) => e.id !== id);
  }
}

export const calendar = new Calendar();

/** A person's short name for avatars and lists. */
export function firstName(u: User | undefined): string {
  if (!u) return '';
  return (u.name || u.email.split('@')[0]).trim().split(/\s+/)[0];
}

/** A stable color for a person's avatar. */
const AVATAR = [
  'bg-rose-500',
  'bg-emerald-500',
  'bg-sky-500',
  'bg-amber-500',
  'bg-fuchsia-500',
  'bg-teal-500',
];
export function avatarColor(id: string): string {
  let h = 0;
  for (const c of id) h = (h * 31 + c.charCodeAt(0)) >>> 0;
  return AVATAR[h % AVATAR.length];
}
