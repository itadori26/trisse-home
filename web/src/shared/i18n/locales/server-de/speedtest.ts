import type { ServerMessageKey } from '../../../api-client/generated/server-messages/speedtest';

/**
 * German translations of the backend error messages of `speedtest` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // An API token with the results:read scope is required.
  srv_d05b6c7e6905729f: 'Ein API-Token mit dem Bereich results:read ist erforderlich.',
  // Speedtest Tracker is not configured yet.
  srv_6f32001a15a15c4d: 'Speedtest Tracker ist noch nicht eingerichtet.',
  // Speedtest Tracker is not reachable.
  srv_cfc88a9871753350: 'Speedtest Tracker ist nicht erreichbar.',
  // Thresholds must be between 0 and 100000 Mbit/s.
  srv_eb4a741553ab9379: 'Schwellenwerte müssen zwischen 0 und 100000 Mbit/s liegen.',
};
