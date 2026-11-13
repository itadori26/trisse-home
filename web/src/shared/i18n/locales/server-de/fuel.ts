import type { ServerMessageKey } from '../../../api-client/generated/server-messages/fuel';

/**
 * German translations of the backend error messages of `fuel` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most {max} favourite stations.
  srv_64c6ae0369f82879: 'Höchstens {max} Lieblingstankstellen.',
  // Choose E5, E10 or Diesel.
  srv_c2ead036ac61a265: 'Wähle E5, E10 oder Diesel.',
  // Could not load fuel prices: {detail}
  srv_3c850049c1f29436: 'Spritpreise konnten nicht geladen werden: {detail}',
  // Every favourite needs a station ID and a name.
  srv_7c971af06beb10fb: 'Jede Lieblingstankstelle braucht eine Tankstellen-ID und einen Namen.',
  // Fuel prices are not configured yet.
  srv_b8ffa62b5ed7e47e: 'Spritpreise sind noch nicht eingerichtet.',
  // Please choose a location.
  srv_4df94a98701c6c20: 'Bitte wähle einen Ort.',
  // The API key must look like 00000000-0000-0000-0000-000000000002.
  srv_90e0ffa88b068f72: 'Der API-Schlüssel muss wie 00000000-0000-0000-0000-000000000002 aussehen.',
  // The fuel price service is not reachable.
  srv_2542e39dc33a11c6: 'Der Spritpreis-Dienst ist nicht erreichbar.',
  // The place name is too long.
  srv_e4facbe924cb2c30: 'Der Ortsname ist zu lang.',
  // The radius must be between 1 and 25 km.
  srv_ad1457baee3bb1c1: 'Der Umkreis muss zwischen 1 und 25 km liegen.',
};
