import type { ServerMessageKey } from '../../../api-client/generated/server-messages/habits';

/**
 * German translations of the backend error messages of `habits` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Only today and yesterday can be checked off.
  srv_a1fc697b9fe6467c: 'Abhaken geht nur für heute und gestern.',
  // The habit does not exist.
  srv_f37ac2a179a60cc1: 'Diese Gewohnheit gibt es nicht.',
};
