import type { ServerMessageKey } from '../../../api-client/generated/server-messages/fritzbox';

/**
 * German translations of the backend error messages of `fritzbox` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Please enter the password of the FRITZ!Box user.
  srv_eabd506dfb319855: 'Bitte gib das Passwort des FRITZ!Box-Benutzers ein.',
  // The FRITZ!Box is not configured yet.
  srv_6f0fbfbef4b05402: 'Die FRITZ!Box ist noch nicht eingerichtet.',
  // The URL must look like http://192.168.178.1:49000.
  srv_d14b64d005d328ef: 'Die URL muss wie http://192.168.178.1:49000 aussehen.',
};
