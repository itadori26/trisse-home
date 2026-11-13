import type { ServerMessageKey } from '../../../api-client/generated/server-messages/transfer';

/**
 * German translations of the backend error messages of `transfer` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Export version {version} is not supported by this version of trisse-home.
  srv_43fe335aacf98b3c: 'Exportversion {version} wird von dieser trisse-home-Version nicht unterstützt.',
  // The data could not be imported: {detail}
  srv_5715790273b7248a: 'Die Daten konnten nicht importiert werden: {detail}',
  // The file is larger than 16 MB.
  srv_e7cb725c138134cc: 'Die Datei ist größer als 16 MB.',
  // The file is not valid JSON.
  srv_b8c878135ee917b1: 'Die Datei ist kein gültiges JSON.',
  // This is not a trisse-home export file.
  srv_d9a79f89c625afc7: 'Das ist keine trisse-home-Exportdatei.',
};
