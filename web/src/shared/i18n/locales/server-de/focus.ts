import type { ServerMessageKey } from '../../../api-client/generated/server-messages/focus';

/**
 * German translations of the backend error messages of `focus` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A session lasts between 1 and 180 minutes.
  srv_e9ad8cf134079d79: 'Eine Einheit dauert zwischen 1 und 180 Minuten.',
  // The session must have ended within the last day.
  srv_addb603b2e8d10d3: 'Die Einheit muss innerhalb des letzten Tages geendet haben.',
};
