import type { ServerMessageKey } from '../../../api-client/generated/server-messages/immich';

/**
 * German translations of the backend error messages of `immich` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Could not connect to Immich: {detail}
  srv_4c9f2919cc2b23d8: 'Verbindung zu Immich fehlgeschlagen: {detail}',
  // Immich is not configured yet.
  srv_b1e270c3a3b1cc7b: 'Immich ist noch nicht eingerichtet.',
  // Immich is not reachable.
  srv_5f51c1adae96a94e: 'Immich ist nicht erreichbar.',
  // The API key is required.
  srv_80ac49019fb5bed8: 'Der API-Schlüssel ist erforderlich.',
  // The URL must look like http://192.168.1.2:2283.
  srv_ce1e99b3a7ab0a0d: 'Die URL muss wie http://192.168.1.2:2283 aussehen.',
  // The thumbnail does not exist.
  srv_6493a23b37be9677: 'Das Vorschaubild existiert nicht.',
};
