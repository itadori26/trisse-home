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
  // Choose a stop for the departures.
  srv_21a0df256f4c1358: 'Wähle eine Haltestelle für die Abfahrten.',
  // Please choose a place with valid coordinates.
  srv_1307228afc94de09: 'Bitte wähle einen Ort mit gültigen Koordinaten.',
};
