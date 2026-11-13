import type { en } from './en';

/** German UI texts of the launcher feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'launcher.title': 'Dienste und Lesezeichen',
  'launcher.filter': 'Dienste und Lesezeichen filtern …',
  'launcher.show': 'Anzeigen',
  'launcher.all': 'Alle',
  'launcher.services': 'Dienste',
  'launcher.bookmarks': 'Lesezeichen',
  'launcher.show_all_services': 'Alle {n} anzeigen',
  'launcher.show_all_bookmarks': 'Alle {n} anzeigen',
  'launcher.no_services': 'Keine passenden Dienste.',
  'launcher.no_bookmarks': 'Keine passenden Lesezeichen.',
  'launcher.online_ms': 'online · {ms} ms',
  'launcher.offline': 'offline',
  'launcher.unknown': 'noch nicht geprüft',
  'launcher.footer_before': 'Was hier erscheint, wählst du in',
  'launcher.footer_link': 'Einstellungen → Darstellung',
  'launcher.footer_ctrl_k': 'Alles ist auch mit Strg K durchsuchbar',
  'launcher.all_online': 'Alle Dienste online',
  'launcher.some_offline': '{n} offline',
};
