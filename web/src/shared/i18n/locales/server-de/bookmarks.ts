import type { ServerMessageKey } from '../../../api-client/generated/server-messages/bookmarks';

/**
 * German translations of the backend error messages of `bookmarks` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A bookmark can have at most 10 tags.
  srv_350d099c21ede260: 'Ein Lesezeichen kann höchstens 10 Tags haben.',
  // At most {max} bookmarks can be imported at once.
  srv_c40eff143d7c40cd: 'Es können höchstens {max} Lesezeichen auf einmal importiert werden.',
  // Enable fetching titles and icons in the bookmark settings first.
  srv_d7bffa395fc89711:
    'Bitte zuerst das Abrufen von Titeln und Icons in den Lesezeichen-Einstellungen aktivieren.',
  // Enable the link check in the bookmark settings first.
  srv_f8405f015011b4d0: 'Bitte zuerst die Linkprüfung in den Lesezeichen-Einstellungen aktivieren.',
  // No bookmarks found. Please export your bookmarks as HTML from the browser.
  srv_2e56580cd044b5c5: 'Keine Lesezeichen gefunden. Bitte exportiere deine Lesezeichen im Browser als HTML.',
  // Tags may only contain letters, digits, - and _ (max. 30 characters).
  srv_7cdf99ade35d99e3: 'Tags dürfen nur Buchstaben, Ziffern, - und _ enthalten (max. 30 Zeichen).',
  // The bookmark does not exist.
  srv_781f5105c04c646b: 'Das Lesezeichen existiert nicht.',
  // The bookmark has no icon.
  srv_0e33c72d2ab4fc96: 'Das Lesezeichen hat kein Icon.',
  // The description must be at most 1000 characters long.
  srv_407e3896b9752ec7: 'Die Beschreibung darf höchstens 1000 Zeichen lang sein.',
  // The file could not be read (max. 8 MB).
  srv_0371382a6f62f53f: 'Die Datei konnte nicht gelesen werden (max. 8 MB).',
  // The folder does not exist.
  srv_9f6aa27e5380b437: 'Der Ordner existiert nicht.',
  // The folder name must be between 1 and 60 characters long.
  srv_c95a6275d4517459: 'Der Ordnername muss zwischen 1 und 60 Zeichen lang sein.',
  // The icon must be a PNG, ICO, JPEG, GIF or WebP image of at most 64 KB.
  srv_c64f54da73ed8e20: 'Das Icon muss ein PNG-, ICO-, JPEG-, GIF- oder WebP-Bild mit höchstens 64 KB sein.',
  // The page could not be read.
  srv_8dec33ac27b43136: 'Die Seite konnte nicht gelesen werden.',
  // The title must be at most 200 characters long.
  srv_1fd35c867febde72: 'Der Titel darf höchstens 200 Zeichen lang sein.',
};
