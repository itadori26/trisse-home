import type { en } from './en';

/** German UI texts of the places feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'settings.section.connections': 'Orte',
  'places.title': 'Orte',
  'places.intro':
    'Gib einen Ort einmal ein und wähle ihn in jedem Widget, das einen Standort braucht: Wetter, Regenradar, Spritpreise und Haltestellen in der Nähe.',
  'places.empty': 'Noch keine Orte gespeichert.',
  'places.used_by': 'Verwendet von {list}',
  'places.unused': 'Von keinem Widget verwendet',
  'places.rename': 'Umbenennen',
  'places.rename_place': '{name} umbenennen',
  'places.remove': 'Entfernen',
  'places.remove_place': '{name} entfernen',
  'places.confirm_remove': '„{name}“ entfernen?',
  'places.confirm_remove_used':
    '„{name}“ wird von {list} verwendet. Diese Widgets behalten ihren Standort, folgen dem Ort aber nicht mehr. Entfernen?',
  'places.add': 'Ort hinzufügen',
  'places.search': 'Suchen',
  'places.search_label': 'Adresse oder Ort suchen',
  'places.search_placeholder': 'z. B. Oldenburg',
  'places.no_results': 'Nichts gefunden.',
  'places.my_location': 'Meinen Standort verwenden',
  'places.locating': 'Standort wird ermittelt …',
  'places.location_denied': 'Der Browser hat den Standort nicht freigegeben.',
  'places.my_location_name': 'Mein Standort',
  'places.name': 'Name',
  'places.save': 'Ort speichern',
  'places.cancel': 'Abbrechen',
  'places.name_empty': 'Gib dem Ort einen Namen.',
  'places.name_too_long': 'Der Name darf höchstens 80 Zeichen haben.',
  'places.coordinates': '{lat}, {lon}',
};
