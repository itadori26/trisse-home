import type { ServerMessageKey } from '../../../api-client/generated/server-messages/backupmonitor';

/**
 * German translations of the backend error messages of `backupmonitor` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The backup source does not exist.
  srv_24356b769e2a8258: 'Die Backup-Quelle existiert nicht.',
  // The grace period must be between 0 minutes and 7 days.
  srv_dcd48aa7c1edb891: 'Die Karenzzeit muss zwischen 0 Minuten und 7 Tagen liegen.',
  // The kind must be backup or check.
  srv_0dbc446d2a829ab3: 'Die Art muss backup oder check sein.',
  // The maximum age must be between 1 hour and 60 days.
  srv_761ff31914abb39c: 'Das Höchstalter muss zwischen 1 Stunde und 60 Tagen liegen.',
  // Use a cron expression with five fields, e.g. 0 3 * * * (daily at 03:00).
  srv_564c3b1fecfb1093:
    'Bitte einen Cron-Ausdruck mit fünf Feldern verwenden, z. B. 0 3 * * * (täglich um 03:00).',
};
