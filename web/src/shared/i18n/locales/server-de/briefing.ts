import type { ServerMessageKey } from '../../../api-client/generated/server-messages/briefing';

/**
 * German translations of the backend error messages of `briefing` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Latitude must be between -90 and 90, longitude between -180 and 180.
  srv_7ef0897648a1c97c:
    'Der Breitengrad muss zwischen -90 und 90 liegen, der Längengrad zwischen -180 und 180.',
  // Notifications are not enabled.
  srv_777fe9a31de38b1d: 'Benachrichtigungen sind nicht aktiviert.',
  // Weekdays are 0 (Sunday) to 6.
  srv_40623194d868a5f0: 'Wochentage sind 0 (Sonntag) bis 6.',
};
