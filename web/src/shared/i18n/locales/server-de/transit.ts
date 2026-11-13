import type { ServerMessageKey } from '../../../api-client/generated/server-messages/transit';

/**
 * German translations of the backend error messages of `transit` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Please choose a stop in the widget settings.
  srv_d92a0962a35bc0f4: 'Bitte wähle in den Widget-Einstellungen eine Haltestelle.',
  // The departure information is not available right now.
  srv_83a3bc209358c090: 'Die Abfahrtsinformationen sind gerade nicht verfügbar.',
};
