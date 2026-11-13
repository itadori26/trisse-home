/** English UI texts of the habits feature (#193). */
export const en = {
  'widget.habits.title': 'Habits',
  'widget.habits.description': 'Daily habits with check-offs, streaks and a 4-week heatmap',
  'habits.new': 'New habit',
  'habits.edit': 'Edit habit',
  'habits.delete': 'Delete',
  'habits.name': 'Name',
  'habits.emoji': 'Emoji',
  'habits.schedule': 'Due on',
  'habits.schedule_hint': 'No weekday selected = every day. Days off do not break a streak.',
  'habits.empty': 'No habits yet.',
  'habits.done_today': '{name} done today',
  'habits.streak_title': 'Current streak (best: {best})',
  'habits.last_weeks': '{name}: the last four weeks',
  'habits.toggle_yesterday': '{name}: check off yesterday',
  'habits.show_heatmap': 'Show the 4-week heatmap',
  'habits.invalid_name': 'The name must be between 1 and 60 characters long.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
