import type { en } from './en';

/** German UI texts of the gamedeals feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.gamedeals.title': 'Spiele-Angebote',
  'widget.gamedeals.description':
    'Rabatte auf deiner Steam-Wunschliste und Gratisspiele im Epic Games Store.',
  'settings.section.gamedeals': 'Spiele-Angebote',
  'gamedeals.connect_under_settings': 'Richte Spiele-Angebote unter Einstellungen → Spiele-Angebote ein.',
  'gamedeals.free': 'Gratis bei Epic',
  'gamedeals.until': 'bis {date}',
  'gamedeals.from': 'ab {date}',
  'gamedeals.wishlist': 'Steam-Wunschliste im Angebot',
  'gamedeals.nothing': 'Gerade keine Angebote.',
  'gamedeals.unreachable': '{source} konnte nicht geladen werden.',
  'gamedeals.steam_id': 'Steam-Profil-URL oder -ID (optional)',
  'gamedeals.steam_hint':
    'Die Wunschliste muss öffentlich sein (Steam → Profil bearbeiten → Privatsphäre → Spieldetails: Öffentlich).',
  'gamedeals.country': 'Land',
  'gamedeals.min_discount': 'Mindestrabatt in %',
  'gamedeals.epic': 'Gratisspiele im Epic Games Store anzeigen',
};
