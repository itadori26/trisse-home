import type { ServerMessageKey } from '../../../api-client/generated/server-messages/readlater';

/**
 * German translations of the backend error messages of `readlater` (#99). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The address is not a web page.
  srv_8eb81e2c9daffa7f: 'Die Adresse ist keine Webseite.',
  // The article does not exist.
  srv_b1c94bd379a22bbd: 'Den Artikel gibt es nicht.',
  // The page is too large.
  srv_573983f536a358ee: 'Die Seite ist zu groß.',
  // The read-later storage is full (200 MB). Delete archived articles first.
  srv_ea46801fada63223:
    'Der Speicher für „Später lesen“ ist voll (200 MB). Lösche zuerst archivierte Artikel.',
  // The status must be unread, read or archived.
  srv_d802b8593fbbdf54: 'Der Status muss ungelesen, gelesen oder archiviert sein.',
};
