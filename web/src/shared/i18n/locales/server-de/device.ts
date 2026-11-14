import type { ServerMessageKey } from '../../../api-client/generated/server-messages/device';

/**
 * German translations of the backend error messages of `device` (#140). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A change of the display mode is still in progress.
  srv_3d12d1ec207507cd: 'Eine Änderung des Anzeigemodus läuft noch.',
  // Choose headless or kiosk.
  srv_5d89741428373d07: 'Wähle „nur Server“ oder „Server mit Anzeige“.',
  // Display modes are only available on the trisse-home appliance image.
  srv_ff97b7a031a91020: 'Anzeigemodi gibt es nur im trisse-home-Image für den Raspberry Pi.',
};
