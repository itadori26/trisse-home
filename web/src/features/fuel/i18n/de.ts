import type { en } from './en';

/** German UI texts of the fuel feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.fuel.title': 'Spritpreise',
  'widget.fuel.description':
    'Günstigste Tankstellen in der Nähe oder deine Lieblingstankstellen (Deutschland, Tankerkönig).',
  'settings.section.fuel': 'Spritpreise',
  'fuel.connect_under_settings': 'Richte Spritpreise unter Einstellungen → Spritpreise ein.',
  'fuel.closed': 'geschlossen',
  'fuel.favourite': '{name} merken',
  'fuel.no_favourites': 'Markiere Tankstellen mit ☆, um sie hier zu sehen.',
  'fuel.no_stations': 'Keine Tankstellen im Umkreis.',
  'fuel.attribution': 'Daten: MTS-K über Tankerkönig (CC BY 4.0)',
  'fuel.view': 'Anzeigen',
  'fuel.view_cheapest': 'Günstigste Tankstellen',
  'fuel.view_favourites': 'Lieblingstankstellen',
  'fuel.search_place': 'Ort suchen',
  'fuel.location': 'Ort: {place} ({lat}, {lon})',
  'fuel.choose_location': 'Bitte wähle einen Ort.',
  'fuel.lat': 'Breitengrad',
  'fuel.lon': 'Längengrad',
  'fuel.fuel': 'Kraftstoff',
  'fuel.radius': 'Umkreis in km',
  'fuel.api_key': 'Tankerkönig-API-Schlüssel {value}',
  'fuel.key_hint':
    'Einen kostenlosen API-Schlüssel gibt es auf creativecommons.tankerkoenig.de. Preise werden höchstens alle fünf Minuten geladen und nur, solange das Widget offen ist.',
  'fuel.invalid_radius': 'Der Umkreis muss zwischen 1 und 25 km liegen.',
};
