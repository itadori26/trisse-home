/** English UI texts of the chore plan feature (#683). */
export const en = {
  'widget.chores.title': 'Chores',
  'widget.chores.description': 'Recurring household tasks with people who take turns.',
  'compact.example.chores': '2 tasks today',
  'chores.new': 'New chore',
  'chores.edit': 'Change chore',
  'chores.title_label': 'Chore',
  'chores.every_label': 'Repeat every … days',
  'chores.every.daily': 'Daily',
  'chores.every.weekly': 'Weekly',
  'chores.every.days': 'Every {n} days',
  'chores.first_due_label': 'First time on',
  'chores.due_label': 'Next time on',
  'chores.assignees_label': 'Who takes turns',
  'chores.assignees_hint': 'One name per line, in order. Each completion passes the chore to the next name.',
  'chores.empty': 'No chores yet.',
  'chores.nothing_today': 'Nothing to do today.',
  'chores.due_heading': 'Today ({n})',
  'chores.later_heading': 'Coming up',
  'chores.today': 'Today',
  'chores.tomorrow': 'Tomorrow',
  'chores.in_days': 'In {n} days',
  'chores.overdue_days': 'Overdue by {n} days',
  'chores.turn': "{name}'s turn",
  'chores.done_named': 'Done: {title}',
  'chores.edit_named': 'Change {title}',
  'chores.remove': 'Remove {title}',
  'chores.error.title': 'Enter a chore of up to 80 characters.',
  'chores.error.every': 'Repeat a chore every 1 to 365 days.',
  'chores.error.due': 'Choose a day.',
  'chores.error.assignees': 'A chore can rotate between at most 8 people.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
