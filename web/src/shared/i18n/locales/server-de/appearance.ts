import type { ServerMessageKey } from '../../../api-client/generated/server-messages/appearance';

/**
 * German translations of the backend error messages of `appearance` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The launcher shows both, services or bookmarks.
  srv_64b6233c48b2792b: 'Der Launcher zeigt beides, nur Dienste oder nur Lesezeichen.',
  // The accent color must look like #3b6ef5.
  srv_1c9c8a99006c73fe: 'Die Akzentfarbe muss wie #3b6ef5 aussehen.',
  // The custom CSS must be at most 64 KiB.
  srv_b410fbcb84d6e168: 'Das eigene CSS darf höchstens 64 KiB groß sein.',
  // The language must be en, de or empty (browser).
  srv_d55a4312396d903f: 'Die Sprache muss en, de oder leer (Browser) sein.',
  // The theme must be system, light or dark.
  srv_d28e94f3f2756a0b: 'Das Design muss system, light oder dark sein.',
  // At most 50 imported themes are supported.
  srv_910076e320aa7e64: 'Es werden höchstens 50 importierte Designs unterstützt.',
  // Built-in themes cannot be deleted.
  srv_23a5dd47efa15b4e: 'Mitgelieferte Designs können nicht gelöscht werden.',
  // Invalid value for {token}.
  srv_fb18e524af59aa63: 'Ungültiger Wert für {token}.',
  // The author must be at most 60 characters long.
  srv_cd1893bbd9659e1f: 'Der Autor darf höchstens 60 Zeichen lang sein.',
  // The file is not a trisse-home theme (JSON).
  srv_68677939789f0976: 'Die Datei ist kein trisse-home-Design (JSON).',
  // The scheme must be light or dark.
  srv_a285cd5903e89e3c: 'Das Farbschema muss light oder dark sein.',
  // The theme does not exist.
  srv_911cd2bd497af69e: 'Das Design gibt es nicht.',
  // The theme id must be 2–40 lowercase letters, digits or '-'.
  srv_02919e323edadf49: "Die Design-ID muss aus 2–40 Kleinbuchstaben, Ziffern oder '-' bestehen.",
  // The theme must set {token}.
  srv_40b907ddd29a3043: 'Das Design muss {token} festlegen.',
  // The theme needs a name of at most 40 characters.
  srv_22e23d9a32c91ddb: 'Das Design braucht einen Namen mit höchstens 40 Zeichen.',
  // There is no custom CSS to restore.
  srv_a7d52a4191c5896d: 'Es gibt kein eigenes CSS zum Wiederherstellen.',
  // Unknown token {token}.
  srv_213674f4c3cbfcf5: 'Unbekanntes Token {token}.',
};
