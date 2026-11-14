import type { en } from './en';

/** German UI texts added to the rss feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'rss.compact': 'Kompakt (nur Titel)',
  'rss.opml_import': 'OPML importieren …',
  'rss.opml_export': 'OPML exportieren',
  'rss.opml_report': '{added} Feeds hinzugefügt, {duplicates} schon abonniert, {invalid} nicht importiert.',
  'rss.opml_too_large': 'Die OPML-Datei ist größer als 2 MB.',
};
