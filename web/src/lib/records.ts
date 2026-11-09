import type { RecordModel, RecordService } from 'pocketbase';
import { pb } from './pb';

// Small helpers for the stores that keep lists of records live.

export function upsert<T extends RecordModel>(
  c: RecordService<T>,
  id: string | null,
  data: Partial<T>,
) {
  return id ? c.update(id, data) : c.create(data);
}

/** The list with a record added, or put in place of its older copy. */
export function put<T extends RecordModel>(list: T[], r: T): T[] {
  return [...list.filter((x) => x.id !== r.id), r];
}

/**
 * Keeps a list up to date with a collection's realtime changes, optionally
 * only the records matching a filter. Resolves to a function that stops it.
 */
export function follow<T extends RecordModel>(
  name: string,
  update: (fn: (list: T[]) => T[]) => void,
  filter?: string,
) {
  return pb.collection<T>(name).subscribe(
    '*',
    ({ action, record }) => {
      update((list) =>
        action === 'delete' ? list.filter((x) => x.id !== record.id) : put(list, record),
      );
    },
    filter ? { filter } : undefined,
  );
}
