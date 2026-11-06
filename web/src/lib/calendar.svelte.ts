import type { RecordModel, RecordService } from 'pocketbase';
import { bySort, lookOf, type Category, type Shown } from './categories';
import type { Doc } from './documents';
import type { Subscription } from './imports';
import { dayOf, type Entry } from './entries';
import { errorMessage, pb, type User } from './pb';

// The household's calendar, its categories, documents and people, kept live:
// a change anyone makes arrives here through PocketBase's realtime
// subscription.
class Calendar {
  entries = $state<Entry[]>([]);
  categories = $state<Category[]>([]);
  documents = $state<Doc[]>([]);
  /** Calendars the household subscribes to by their link. */
  subscriptions = $state<Subscription[]>([]);
  people = $state<User[]>([]);
  loaded = $state(false);
  error = $state('');
  today = $state(dayOf(new Date()));

  #stop: (() => void) | null = null;

  async start() {
    try {
      const [entries, categories, documents, people, subscriptions] = await Promise.all([
        pb.collection<Entry>('entries').getFullList({ sort: 'start', batch: 500 }),
        pb.collection<Category>('categories').getFullList({ sort: 'sort,created' }),
        pb.collection<Doc>('documents').getFullList({ sort: 'title' }),
        pb.collection<User>('users').getFullList({ sort: 'name' }),
        pb.collection<Subscription>('calendars').getFullList({ sort: 'created' }),
      ]);
      this.entries = entries;
      this.categories = categories;
      this.documents = documents;
      this.people = people;
      this.subscriptions = subscriptions;
      this.loaded = true;
    } catch (err) {
      this.error = errorMessage(err, "Couldn't load the calendar.");
      return;
    }

    const stops = await Promise.all([
      follow<Entry>('entries', (fn) => (this.entries = fn(this.entries))),
      follow<Category>('categories', (fn) => (this.categories = fn(this.categories))),
      follow<Doc>('documents', (fn) => (this.documents = fn(this.documents))),
      follow<Subscription>('calendars', (fn) => (this.subscriptions = fn(this.subscriptions))),
    ]);
    // the day changes at midnight, and when the phone wakes the app up
    const tick = () => (this.today = dayOf(new Date()));
    const timer = setInterval(tick, 60_000);
    document.addEventListener('visibilitychange', tick);
    this.#stop = () => {
      stops.forEach((stop) => stop());
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

  /** How an entry shows: its category's name, color and icon. */
  look(e: Pick<Entry, 'kind' | 'category'>): Shown {
    return lookOf(e, this.categories);
  }

  get sortedCategories(): Category[] {
    return [...this.categories].sort(bySort);
  }

  async save(id: string | null, data: Partial<Entry>): Promise<Entry> {
    const saved = await upsert(pb.collection<Entry>('entries'), id, data);
    // show it right away; the realtime event that follows is the same record
    this.entries = put(this.entries, saved);
    return saved;
  }

  async remove(id: string) {
    await pb.collection('entries').delete(id);
    this.entries = this.entries.filter((e) => e.id !== id);
  }

  async saveCategory(id: string | null, data: Partial<Category>): Promise<Category> {
    const saved = await upsert(pb.collection<Category>('categories'), id, data);
    this.categories = put(this.categories, saved);
    return saved;
  }

  async removeCategory(id: string) {
    await pb.collection('categories').delete(id);
    this.categories = this.categories.filter((c) => c.id !== id);
  }

  async saveDocument(id: string | null, data: FormData | Partial<Doc>): Promise<Doc> {
    const docs = pb.collection<Doc>('documents');
    const saved = id ? await docs.update(id, data) : await docs.create(data);
    this.documents = put(this.documents, saved);
    return saved;
  }

  /** Puts a document the server changed (shared or not) in place. */
  async reloadDocument(id: string) {
    this.documents = put(this.documents, await pb.collection<Doc>('documents').getOne(id));
  }

  /** Puts a subscription the server changed (added or refreshed) in place. */
  putSubscription(s: Subscription) {
    this.subscriptions = put(this.subscriptions, s);
  }

  /** Stops a subscription; its entries go with it. */
  async removeSubscription(id: string) {
    await pb.collection('calendars').delete(id);
    this.subscriptions = this.subscriptions.filter((s) => s.id !== id);
    this.entries = this.entries.filter((e) => e.calendar !== id);
  }

  async removeDocument(id: string) {
    await pb.collection('documents').delete(id);
    this.documents = this.documents.filter((d) => d.id !== id);
  }
}

function upsert<T extends RecordModel>(c: RecordService<T>, id: string | null, data: Partial<T>) {
  return id ? c.update(id, data) : c.create(data);
}

function put<T extends RecordModel>(list: T[], r: T): T[] {
  return [...list.filter((x) => x.id !== r.id), r];
}

/** Keeps a list up to date with a collection's realtime changes. */
function follow<T extends RecordModel>(name: string, update: (fn: (list: T[]) => T[]) => void) {
  return pb.collection<T>(name).subscribe('*', ({ action, record }) => {
    update((list) =>
      action === 'delete' ? list.filter((x) => x.id !== record.id) : put(list, record),
    );
  });
}

export const calendar = new Calendar();

/** A person's short name for avatars and lists. */
export function firstName(u: User | undefined): string {
  if (!u) return '';
  // other people's email isn't sent, so someone without a name has neither
  const name = (u.name || u.email?.split('@')[0] || '').trim();
  return name ? name.split(/\s+/)[0] : 'Someone';
}

/** A stable color for a person's avatar. */
const AVATAR = [
  'bg-rose-100 text-rose-800 dark:bg-rose-400/20 dark:text-rose-200',
  'bg-emerald-100 text-emerald-800 dark:bg-emerald-400/20 dark:text-emerald-200',
  'bg-sky-100 text-sky-800 dark:bg-sky-400/20 dark:text-sky-200',
  'bg-amber-100 text-amber-800 dark:bg-amber-400/20 dark:text-amber-200',
  'bg-violet-100 text-violet-800 dark:bg-violet-400/20 dark:text-violet-200',
  'bg-teal-100 text-teal-800 dark:bg-teal-400/20 dark:text-teal-200',
  'bg-orange-100 text-orange-800 dark:bg-orange-400/20 dark:text-orange-200',
  'bg-lime-100 text-lime-800 dark:bg-lime-400/20 dark:text-lime-200',
];
export function avatarColor(id: string): string {
  let h = 0;
  for (const c of id) h = (h * 31 + c.charCodeAt(0)) >>> 0;
  return AVATAR[h % AVATAR.length];
}
