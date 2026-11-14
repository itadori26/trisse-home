import type { en } from './en';

/** German UI texts of the space feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.space.title': 'Raumfahrt',
  'widget.space.description':
    'Nächste Raketenstarts (Launch Library 2) und das NASA-Astronomiebild des Tages.',
  'settings.section.space': 'Raumfahrt',
  'space.intro':
    'Beide Quellen sind aus, bis du sie einschaltest; danach werden sie nur geladen, solange ein Widget sie zeigt.',
  'space.launches_on': 'Raketenstarts anzeigen (Launch Library 2)',
  'space.picture_on': 'Astronomiebild des Tages anzeigen (NASA)',
  'space.nasa_key': 'NASA-API-Schlüssel (optional)',
  'space.nasa_key_stored': 'gespeichert – leer lassen, um ihn zu behalten',
  'space.nasa_key_hint':
    'Ohne eigenen kostenlosen Schlüssel von api.nasa.gov wird der gemeinsame DEMO_KEY genutzt, der pro IP-Adresse stark begrenzt ist.',
  'space.background': 'Bild des Tages als Hintergrund',
  'space.background_hint':
    'Nutze diese Adresse als Tab-Hintergrund (Bearbeiten → 🖼) oder unter Erscheinungsbild. Sie zeigt immer das heutige Bild, geladen von diesem Server.',
  'space.background_copy': 'Hintergrund-Adresse kopieren',
  'space.saved': 'Gespeichert.',
  'space.show': 'Anzeigen',
  'space.show_both': 'Starts und Bild',
  'space.show_launches': 'Nur Starts',
  'space.show_picture': 'Nur Bild',
  'space.count': 'Starts: {n}',
  'space.switched_off': 'Schalte die Raumfahrt-Quellen unter Einstellungen → Raumfahrt ein.',
  'space.next_launches': 'Nächste Starts',
  'space.no_launches': 'Keine Starts angekündigt.',
  'space.now': 'jetzt',
  'space.in_minutes': 'in {n} Min.',
  'space.in_hours': 'in {n} Std.',
  'space.in_days': 'in {n} Tagen',
  'space.live': 'Live',
  'space.status_go': 'Go',
  'space.status_tbd': 'offen',
  'space.status_tbc': 'unbestätigt',
  'space.status_hold': 'Hold',
  'space.status_in_flight': 'im Flug',
  'space.status_success': 'Erfolg',
  'space.status_failure': 'Fehlschlag',
  'space.status_partial_failure': 'Teilerfolg',
  'space.status_unknown': '?',
  'space.picture': 'Bild des Tages',
  'space.video_today': 'Das heutige Bild ist ein Video.',
  'space.open': 'Öffnen',
  'space.more': 'Mehr',
};
