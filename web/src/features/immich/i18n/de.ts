import type { en } from './en';

/** German UI texts of the immich feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.immich.title': 'Immich',
  'widget.immich.description': 'Statistik der Fotobibliothek und Erinnerungen des Tages.',
  'settings.section.immich': 'Immich',
  'immich.connect_under_settings': 'Verbinde Immich unter Einstellungen → Immich.',
  'immich.photos': 'Fotos',
  'immich.videos': 'Videos',
  'immich.storage': 'Speicher',
  'immich.years_ago': 'Heute vor {n} Jahren',
  'immich.memory_photo': 'Foto von vor {n} Jahren',
  'immich.show_memories': 'Erinnerungen des Tages anzeigen',
  'immich.url': 'Immich-URL',
  'immich.api_key': 'API-Schlüssel {value}',
  'immich.key_hint':
    'Lege einen API-Schlüssel unter Kontoeinstellungen → API-Schlüssel mit den Berechtigungen asset.statistics, server.storage, memory.read und asset.view an. Er wird verschlüsselt gespeichert; Fotos lädt trisse-home selbst.',
};
