import type { ServerMessageKey } from '../../../api-client/generated/server-messages/space';

/**
 * German translations of the backend error messages of `space` (#64). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The NASA API key looks wrong.
  srv_1725722a41c512af: 'Der NASA-API-Schlüssel sieht falsch aus.',
  // The picture is not hosted by NASA.
  srv_374c6d77e754cb26: 'Das Bild liegt nicht bei der NASA.',
  // The space data is not available right now.
  srv_21a524286ae6e520: 'Die Raumfahrtdaten sind gerade nicht verfügbar.',
  // This space source is switched off in the settings.
  srv_1b5d999910a8dd39: 'Diese Raumfahrt-Quelle ist in den Einstellungen ausgeschaltet.',
  // Today's picture is not an image.
  srv_9825ec9eef300ecf: 'Das heutige Bild ist kein Bild (z. B. ein Video).',
};
