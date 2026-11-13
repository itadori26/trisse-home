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
};
