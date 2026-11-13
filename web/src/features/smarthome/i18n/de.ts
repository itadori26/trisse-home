import type { en } from './en';

/** German UI texts of the smarthome feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.smarthome.title': 'Smart Home',
  'widget.smarthome.description':
    'Zustände von Home-Assistant-Geräten und -Sensoren; optional Schalter und Taster.',
  'settings.section.smarthome': 'Home Assistant',
  'smarthome.connect_under_settings': 'Verbinde Home Assistant unter Einstellungen → Home Assistant.',
  'smarthome.choose_entities': 'Wähle Entitäten in den Widget-Einstellungen.',
  'smarthome.unavailable': 'nicht verfügbar',
  'smarthome.on': 'an',
  'smarthome.off': 'aus',
  'smarthome.turn_on': 'Einschalten',
  'smarthome.turn_off': 'Ausschalten',
  'smarthome.press': 'Drücken',
  'smarthome.confirm_toggle': '{name} umschalten?',
  'smarthome.confirm_press': '{name} drücken?',
  'smarthome.search_entities': 'Entitäten suchen',
  'smarthome.selected': '{count} von höchstens {max} ausgewählt',
  'smarthome.url': 'Home-Assistant-URL',
  'smarthome.token': 'Langlebiges Zugriffstoken {value}',
  'smarthome.token_hint':
    'Lege das Token in Home Assistant unter Profil → Sicherheit → Langlebige Zugriffstoken an. Es wird verschlüsselt gespeichert und nie wieder angezeigt.',
  'smarthome.actions_hint':
    'Schalter und Taster bleiben schreibgeschützt, solange SMARTHOME_ACTIONS_ENTITIES ihre Entitäts-IDs nicht auflistet.',
};
