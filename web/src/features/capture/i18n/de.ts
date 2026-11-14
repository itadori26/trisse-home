import type { en } from './en';

/** German UI texts of the quick capture page; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'capture.title': 'In trisse-home speichern',
  'capture.as': 'Speichern als',
  'capture.bookmark': 'Lesezeichen',
  'capture.todo': 'Aufgabe',
  'capture.note': 'Notiz',
  'capture.name': 'Titel',
  'capture.url': 'Link',
  'capture.text': 'Text',
  'capture.save': 'Speichern',
  'capture.missing': 'Gib einen Titel ein (bzw. einen gültigen Link für ein Lesezeichen).',
  'capture.saved_bookmark': 'Lesezeichen gespeichert.',
  'capture.saved_todo': 'Aufgabe gespeichert.',
  'capture.saved_note': 'Notiz gespeichert.',
  'capture.undo': 'Rückgängig',
  'capture.undone': 'Wieder entfernt.',
  'capture.close': 'Schließen',
  'capture.dashboard': 'Dashboard öffnen',
  'capture.another': 'Noch etwas speichern',
  'capture.bookmarklet': 'Schnell speichern von jeder Seite',
  'capture.bookmarklet_hint':
    'Zieh diesen Link in deine Lesezeichenleiste. Auf jeder Seite öffnet er ein kleines Fenster, das die Seite als Lesezeichen, Aufgabe oder Notiz speichert (markierter Text wird mitgenommen). Auf dem Handy teilst du stattdessen an die installierte App.',
  'capture.bookmarklet_link': '＋ trisse-home',
  'capture.bookmarklet_copy': 'Bookmarklet kopieren',
};
