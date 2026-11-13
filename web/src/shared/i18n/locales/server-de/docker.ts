import type { ServerMessageKey } from '../../../api-client/generated/server-messages/docker';

/**
 * German translations of the backend error messages of `docker` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A password or access token is required for new registries.
  srv_865d0fdde8c133ca: 'Für neue Registries ist ein Passwort oder Zugriffstoken erforderlich.',
  // Actions are not enabled for this container (DOCKER_ACTIONS / DOCKER_ACTIONS_CONTAINERS).
  srv_fc1c5d36ee0973d3:
    'Aktionen sind für diesen Container nicht freigeschaltet (DOCKER_ACTIONS / DOCKER_ACTIONS_CONTAINERS).',
  // At most 20 registries can be stored.
  srv_2ad8dffbd734017a: 'Es können höchstens 20 Registries gespeichert werden.',
  // Each registry may only be listed once.
  srv_4ae76664f7b46c6b: 'Jede Registry darf nur einmal eingetragen sein.',
  // Each registry needs a host name (e.g. ghcr.io) and a user name.
  srv_e9bcefee5bd2a359: 'Jede Registry braucht einen Hostnamen (z. B. ghcr.io) und einen Benutzernamen.',
  // The Docker API is not reachable. Does the socket proxy allow IMAGES=1?
  srv_1120b37481579f63: 'Die Docker-API ist nicht erreichbar. Erlaubt der Socket-Proxy IMAGES=1?',
  // The image update check is disabled. Enable it in Settings → Docker updates.
  srv_c9b09323ab5702c6:
    'Die Prüfung auf Image-Updates ist ausgeschaltet. Sie lässt sich unter Einstellungen → Docker-Updates einschalten.',
};
