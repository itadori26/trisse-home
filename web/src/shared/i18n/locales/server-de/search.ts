import type { ServerMessageKey } from '../../../api-client/generated/server-messages/search';

/**
 * German translations of the backend error messages of `search` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A shortcut with this keyword already exists.
  srv_76a5625367cbe997: 'Es gibt bereits ein Kürzel mit diesem Schlüsselwort.',
  // The URL must contain {query}.
  srv_1605df454fade766: 'Die URL muss {query} enthalten.',
  // The keyword must consist of 1 to 16 letters or digits.
  srv_d6e09729bc0d1f3b: 'Das Schlüsselwort muss aus 1 bis 16 Buchstaben oder Ziffern bestehen.',
  // The search query is too long.
  srv_657058c57c5faf46: 'Die Suchanfrage ist zu lang.',
  // The search query must not be empty.
  srv_b828d9f84cec079a: 'Die Suchanfrage darf nicht leer sein.',
  // The shortcut does not exist.
  srv_2b29e30433c7ad15: 'Das Kürzel existiert nicht.',
  // Unknown search provider.
  srv_dea670a71d2295ca: 'Unbekannte Suchmaschine.',
};
