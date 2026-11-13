import type { ServerMessageKey } from '../../../api-client/generated/server-messages/smarthome';

/**
 * German translations of the backend error messages of `smarthome` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A long-lived access token is required.
  srv_084af9ca0c7d167d: 'Ein langlebiges Zugriffstoken ist erforderlich.',
  // Actions are not enabled for this entity (SMARTHOME_ACTIONS_ENTITIES).
  srv_0f4d7732f2cd4899: 'Aktionen sind für diese Entität nicht aktiviert (SMARTHOME_ACTIONS_ENTITIES).',
  // At most {max} entities per widget.
  srv_3a4c7bf887ab6487: 'Höchstens {max} Entitäten pro Widget.',
  // Could not connect to Home Assistant: {detail}
  srv_789dca50a23c7cb8: 'Verbindung zu Home Assistant fehlgeschlagen: {detail}',
  // Home Assistant is not configured yet.
  srv_bcbabc6a08a54107: 'Home Assistant ist noch nicht eingerichtet.',
  // Home Assistant is not reachable.
  srv_dc90aa2a1e0a5f62: 'Home Assistant ist nicht erreichbar.',
  // Home Assistant rejected the action: {detail}
  srv_af1423afec138ad6: 'Home Assistant hat die Aktion abgelehnt: {detail}',
  // The URL must look like http://homeassistant.local:8123.
  srv_d26f0cb52cfa0948: 'Die URL muss wie http://homeassistant.local:8123 aussehen.',
  // The entity does not exist.
  srv_2ef3700ca8a09eac: 'Die Entität existiert nicht.',
  // This entity has no action; only switches and buttons can be used.
  srv_a6715a41424b56f8: 'Diese Entität hat keine Aktion; nur Schalter und Taster lassen sich bedienen.',
};
