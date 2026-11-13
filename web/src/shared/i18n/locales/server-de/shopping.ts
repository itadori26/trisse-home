import type { ServerMessageKey } from '../../../api-client/generated/server-messages/shopping';

/**
 * German translations of the backend error messages of `shopping` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Items must be between 1 and 200 characters long.
  srv_593a1c118a3d9d67: 'Einträge müssen zwischen 1 und 200 Zeichen lang sein.',
  // The item does not exist (any more).
  srv_77686a482211f04a: 'Der Eintrag existiert nicht (mehr).',
  // The item was changed in another app at the same time. Please try again.
  srv_7faea27f96c82f5f:
    'Der Eintrag wurde gleichzeitig in einer anderen App geändert. Bitte erneut versuchen.',
  // The shopping list is not configured yet.
  srv_7d7ee048ea53fd65: 'Die Einkaufsliste ist noch nicht eingerichtet.',
};
