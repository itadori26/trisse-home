import type { en } from './en';

/** German UI texts of the habits feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.habits.title': 'Gewohnheiten',
  'widget.habits.description': 'Tägliche Gewohnheiten abhaken, mit Serien und 4-Wochen-Übersicht',
  'habits.new': 'Neue Gewohnheit',
  'habits.edit': 'Gewohnheit bearbeiten',
  'habits.delete': 'Löschen',
  'habits.name': 'Name',
  'habits.emoji': 'Emoji',
  'habits.schedule': 'Fällig am',
  'habits.schedule_hint': 'Kein Wochentag gewählt = jeden Tag. Freie Tage unterbrechen keine Serie.',
  'habits.empty': 'Noch keine Gewohnheiten.',
  'habits.done_today': '{name} heute erledigt',
  'habits.streak_title': 'Aktuelle Serie (beste: {best})',
  'habits.last_weeks': '{name}: die letzten vier Wochen',
  'habits.toggle_yesterday': '{name}: gestern abhaken',
  'habits.show_heatmap': '4-Wochen-Übersicht anzeigen',
  'habits.invalid_name': 'Der Name muss zwischen 1 und 60 Zeichen lang sein.',
};
