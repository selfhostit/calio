import type { RecordModel } from 'pocketbase';

// The shop: outfits for companions, bought with berries. The server has the
// catalog and the prices (backend/play/shop); each item is drawn by its id.

export type Slot = 'head' | 'neck' | 'face';

export interface Item {
  id: string;
  name: string;
  slot: Slot;
  price: number;
}

/** Something someone bought, and whether their companion is wearing it. */
export interface Owned extends RecordModel {
  user: string;
  item: string;
  equipped: boolean;
}

/** How grown a companion is: it grows with the habits done over time. */
export type Stage = 'cub' | 'young' | 'grown';

export const STAGES: Record<Stage, string> = { cub: 'Cub', young: 'Young', grown: 'Grown-up' };

/** Your berries, streak and how grown your companion is (GET /api/app/play/me). */
export interface Progress {
  berries: number;
  streak: number;
  lifetime: number;
  stage: Stage;
  /** Habits done at which it grows next; 0 when it's grown up. */
  next_stage: number;
}

/** The ids of what's being worn. */
export const wearing = (owned: Owned[]) => owned.filter((o) => o.equipped).map((o) => o.item);
