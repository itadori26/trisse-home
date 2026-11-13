import type { ServerMessageKey } from '../../../api-client/generated/server-messages/timeline';

/**
 * German translations of the backend error messages of `timeline` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The retention must be between 1 and 365 days.
  srv_2367e29c6edb5841: 'Die Aufbewahrungsdauer muss zwischen 1 und 365 Tagen liegen.',
};
