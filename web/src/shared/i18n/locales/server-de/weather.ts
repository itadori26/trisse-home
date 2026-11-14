import type { ServerMessageKey } from '../../../api-client/generated/server-messages/weather';

/**
 * German translations of the backend error messages of `weather` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Latitude and longitude are required.
  srv_9d350adc128ef832: 'Breiten- und Längengrad sind erforderlich.',
  // Please enter between 2 and 100 characters.
  srv_990b607e23145ec7: 'Bitte zwischen 2 und 100 Zeichen eingeben.',
  // The latitude must be between -90 and 90.
  srv_3f60e80ed2350494: 'Der Breitengrad muss zwischen -90 und 90 liegen.',
  // The longitude must be between -180 and 180.
  srv_39566e8cbb1ec35d: 'Der Längengrad muss zwischen -180 und 180 liegen.',
  // The rain radar is only available for Germany.
  srv_d5d2cea1e1d1fa8d: 'Das Regenradar ist nur für Deutschland verfügbar.',
  // The weather warnings are not available right now.
  srv_314a73e152936a17: 'Die Wetterwarnungen sind gerade nicht verfügbar.',
  // The pollen and UV forecasts are not available right now.
  srv_cb30f7793468eebc: 'Pollenflug und UV-Index sind gerade nicht verfügbar.',
};
