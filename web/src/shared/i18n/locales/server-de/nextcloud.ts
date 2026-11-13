import type { ServerMessageKey } from '../../../api-client/generated/server-messages/nextcloud';

/**
 * German translations of the backend error messages of `nextcloud` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Nextcloud is not configured yet.
  srv_852391e96f8a11b2: 'Nextcloud ist noch nicht eingerichtet.',
};
