import type { en } from './en';

/** German UI texts of the read-later feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.readlater.title': 'Später lesen',
  'widget.readlater.description':
    'Gespeicherte Artikel mit einer sauberen Lesefassung, die lesbar bleibt, wenn die Seite weg ist.',
  'settings.section.readlater': 'Später lesen',
  'readlater.pending': 'Lesekopie folgt im Hintergrund',
  'readlater.more': 'Aufräumen und Import',
  'readlater.archive_days': 'Archivierte Artikel löschen',
  'readlater.after_days': 'nach {n} Tagen',
  'readlater.keep': 'nie',
  'readlater.import_pocket': 'Pocket-Export importieren …',
  'readlater.import_hint':
    'Die CSV-Datei aus dem Pocket-Export. Die Links werden sofort übernommen; Lesekopien folgen im Hintergrund, einige alle fünf Minuten.',
  'readlater.import_report':
    '{added} hinzugefügt, {duplicates} schon gespeichert, {invalid} nicht importierbar.',
  'readlater.import_too_large': 'Die Exportdatei darf höchstens 5 MB groß sein.',
  'readlater.read_later': 'Später lesen',
  'readlater.read_later_item': '{title} später lesen',
  'readlater.saved_for_later': 'Für später gespeichert',
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
