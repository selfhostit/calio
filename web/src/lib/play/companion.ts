import type { RecordModel } from 'pocketbase';

// Everyone's companion: an animal of their own whose mood follows their
// habits. The red panda is the first kind; each kind is drawn by the web app
// (components/play/Companion.svelte).

export type Species = 'red_panda';

export interface Companion extends RecordModel {
  user: string;
  name: string;
  species: Species;
}

export const SPECIES: Record<Species, { label: string; name: string }> = {
  red_panda: { label: 'Red panda', name: 'Ember' },
};
