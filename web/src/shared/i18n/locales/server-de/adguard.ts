import type { ServerMessageKey } from '../../../api-client/generated/server-messages/adguard';

/**
 * German translations of the backend error messages of `adguard` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Username and password are required.
  srv_ed3a053401c9b56a: 'Benutzername und Passwort sind erforderlich.',
  // The DNS filter is not reachable.
  srv_4a327952321d8a3f: 'Der DNS-Filter ist nicht erreichbar.',
  // This DNS filter is not configured yet.
  srv_fcbb5bbe636e4559: 'Dieser DNS-Filter ist noch nicht eingerichtet.',
  // Unknown DNS filter, use adguard or pihole.
  srv_1a87cc7977ed4993: 'Unbekannter DNS-Filter, verwende adguard oder pihole.',
  // Pause for 5, 30 or 60 minutes.
  srv_0c0b1a4d7e09267c: 'Pausiere für 5, 30 oder 60 Minuten.',
  // Pausing is switched off for this DNS filter. Allow it with DNS_PAUSE_PROVIDERS on the server.
  srv_6af65a6080fe6b52:
    'Pausieren ist für diesen DNS-Filter ausgeschaltet. Erlaube es mit DNS_PAUSE_PROVIDERS auf dem Server.',
};
