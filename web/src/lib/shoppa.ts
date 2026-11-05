import { pb } from './pb';

// Shoppa, the household's shopping list app, when the server is set up to
// reach it (SHOPPA_URL and a worker account; see docs/deploy.md).

let enabled: Promise<boolean> | null = null;

/** Whether shopping lists can be sent to Shoppa; asked once. */
export function shoppaEnabled(): Promise<boolean> {
  enabled ??= pb
    .send<{ enabled: boolean }>('/api/app/shoppa', {})
    .then((r) => r.enabled)
    .catch(() => false);
  return enabled;
}

/** Sends what's still to buy on an entry's shopping list to Shoppa. */
export function sendToShoppa(entryId: string): Promise<{ sent: number; list: string }> {
  return pb.send(`/api/app/shoppa/send/${entryId}`, { method: 'POST' });
}
