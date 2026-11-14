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
  // Archived articles can be kept for 1 to 365 days, or forever (0).
  srv_cc69e38143dedd78: 'Archivierte Artikel können 1 bis 365 Tage oder für immer (0) behalten werden.',
  // At most 2000 links can be imported at once.
  srv_03b5b5f039e06831: 'Es können höchstens 2000 Links auf einmal importiert werden.',
  // The export file must be at most 5 MB.
  srv_22fa6fd55903ce2e: 'Die Exportdatei darf höchstens 5 MB groß sein.',
  // The file is not a Pocket export (CSV with title, url, time_added, status).
  srv_8680771d11b72ed2: 'Die Datei ist kein Pocket-Export (CSV mit title, url, time_added, status).',
};
