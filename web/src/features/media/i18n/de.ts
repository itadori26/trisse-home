import type { en } from './en';

/** German UI texts of the media feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.media.title': 'Medienserver',
  'widget.media.description': 'Was gerade läuft und was neu ist auf Jellyfin oder Plex.',
  'settings.section.media': 'Medienserver',
  'media.connect_under_settings': 'Verbinde Jellyfin oder Plex unter Einstellungen → Medienserver.',
  'media.now_playing': 'Läuft gerade',
  'media.nothing_playing': 'Gerade läuft nichts.',
  'media.recently_added': 'Neu hinzugefügt',
  'media.server': 'Medienserver',
  'media.url': '{name}-URL',
  'media.jellyfin_key': 'API-Schlüssel {value}',
  'media.plex_token': 'Plex-Token {value}',
  'media.jellyfin_hint':
    'Lege einen API-Schlüssel in Jellyfin unter Dashboard → API-Schlüssel an. Er wird verschlüsselt gespeichert; Cover lädt trisse-home selbst.',
  'media.plex_hint':
    'Nutze das X-Plex-Token deines Kontos (Plex-Hilfeartikel „Finding an authentication token“). Es wird verschlüsselt gespeichert; Cover lädt trisse-home selbst.',
};
