import type { ServerMessageKey } from '../../../api-client/generated/server-messages/diskhealth';

/**
 * German translations of the backend error messages of `diskhealth` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // No S.M.A.R.T. export found. Run scripts/smart-export.sh on the host (see README).
  srv_74822a0f2779c188:
    'Kein S.M.A.R.T.-Export gefunden. Bitte scripts/smart-export.sh auf dem Host ausführen (siehe README).',
};
