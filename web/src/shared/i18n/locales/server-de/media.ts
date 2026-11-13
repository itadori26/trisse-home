import type { ServerMessageKey } from '../../../api-client/generated/server-messages/media';

/**
 * German translations of the backend error messages of `media` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose Jellyfin or Plex.
  srv_52395ea57eb8771f: 'Wähle Jellyfin oder Plex.',
  // Could not connect to the media server: {detail}
  srv_65d6b40b9e5a9159: 'Verbindung zum Medienserver fehlgeschlagen: {detail}',
  // The API key or token is required.
  srv_2aa10686fa162dde: 'Der API-Schlüssel oder das Token ist erforderlich.',
  // The URL must look like http://192.168.1.2:8096.
  srv_ec3846ce5d9b8caf: 'Die URL muss wie http://192.168.1.2:8096 aussehen.',
  // The media server is not configured yet.
  srv_7b276d2193d07bc8: 'Der Medienserver ist noch nicht eingerichtet.',
  // The media server is not reachable.
  srv_4c92147e0244306f: 'Der Medienserver ist nicht erreichbar.',
  // The poster does not exist.
  srv_2b05bbdb56e575c6: 'Das Cover existiert nicht.',
};
