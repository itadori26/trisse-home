import type { ServerMessageKey } from '../../../api-client/generated/server-messages/arr';

/**
 * German translations of the backend error messages of `arr` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most {max} instances are supported.
  srv_e087fafa20933ecf: 'Höchstens {max} Instanzen werden unterstützt.',
  // Choose Sonarr, Radarr or Prowlarr.
  srv_35314e09732db39d: 'Wähle Sonarr, Radarr oder Prowlarr.',
  // Could not connect to {app}: {detail}
  srv_b4411af5f4818c54: 'Verbindung zu {app} fehlgeschlagen: {detail}',
  // No Sonarr, Radarr or Prowlarr instance is configured yet.
  srv_e7b97389eba1653e: 'Noch keine Sonarr-, Radarr- oder Prowlarr-Instanz eingerichtet.',
  // The API key is required (Settings → General).
  srv_7a0dfb1a22ec0b96: 'Der API-Schlüssel ist erforderlich (Settings → General).',
  // The URL must look like http://192.168.1.2:8989.
  srv_f850721911b29bc6: 'Die URL muss wie http://192.168.1.2:8989 aussehen.',
  // The instance does not exist.
  srv_c8e8fb2032a453cc: 'Die Instanz existiert nicht.',
};
