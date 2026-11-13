import type { ServerMessageKey } from '../../../api-client/generated/server-messages/holidays';

/**
 * German translations of the backend error messages of `holidays` (#55). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Could not load the school holidays: {detail}
  srv_e204fc52da15fb72: 'Die Schulferien konnten nicht geladen werden: {detail}',
};
