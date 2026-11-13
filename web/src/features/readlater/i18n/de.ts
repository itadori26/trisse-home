import type { en } from './en';

/** German UI texts of the read-later feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.readlater.title': 'Später lesen',
  'widget.readlater.description':
    'Gespeicherte Artikel mit einer sauberen Lesefassung, die lesbar bleibt, wenn die Seite weg ist.',
  'settings.section.readlater': 'Später lesen',
  'readlater.intro':
    'Gespeicherte Seiten bekommen eine saubere Lesefassung mit ihren Bildern, gespeichert auf diesem Server. Angefragt wird nur die Seite, die du speicherst, und zwar einmal.',
  'readlater.save': 'Speichern',
  'readlater.save_label': 'Adresse der Seite zum späteren Lesen',
  'readlater.save_placeholder': '+ Später lesen: Link einfügen',
  'readlater.saving': 'Seite wird gespeichert …',
  'readlater.url_invalid': 'Gib die http(s)-Adresse einer Webseite ein.',
  'readlater.empty': 'Nichts zu lesen.',
  'readlater.minutes': '{n} Min.',
  'readlater.saved_on': 'gespeichert am {date}',
  'readlater.original': 'Original',
  'readlater.no_copy': 'Von dieser Seite gibt es keine Lesefassung.',
  'readlater.archive': 'Archivieren',
  'readlater.mark_read': '„{title}“ als gelesen markieren',
  'readlater.archive_item': '„{title}“ archivieren',
  'readlater.restore_item': 'Wieder ungelesen: „{title}“',
  'readlater.delete_item': '„{title}“ löschen',
  'readlater.delete_confirm': '„{title}“ mit Lesefassung löschen?',
  'readlater.status': 'Anzeigen',
  'readlater.status_unread': 'Ungelesen',
  'readlater.status_read': 'Gelesen',
  'readlater.status_archived': 'Archiviert',
  'readlater.storage': 'Diese Artikel belegen {size}.',
  'readlater.count': 'Artikel: {n}',
};
