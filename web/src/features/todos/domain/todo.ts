/** A to-do item. `due` is a date string (YYYY-MM-DD) or ''. */
export interface Todo {
  id: string;
  title: string;
  due: string;
  done: boolean;
  overdue: boolean;
}

/** Configuration of the to-do widget. */
export interface TodosConfig {
  showDone: boolean;
}

/** Reads a stored widget configuration. */
export function parseTodosConfig(raw: Record<string, unknown>): TodosConfig {
  return { showDone: raw['showDone'] === true };
}

/**
 * Formats a due date relative to today ("Today", "Tomorrow", weekday within
 * a week, otherwise the date).
 */
export function formatDue(due: string, today: Date, locale?: string): string {
  if (!due) return '';
  const [y, m, d] = due.split('-').map(Number);
  if (!y || !m || !d) return due;
  const date = new Date(y, m - 1, d);
  const start = new Date(today.getFullYear(), today.getMonth(), today.getDate());
  const days = Math.round((date.getTime() - start.getTime()) / 86_400_000);
  if (days === 0) return 'Today';
  if (days === 1) return 'Tomorrow';
  if (days === -1) return 'Yesterday';
  if (days > 1 && days < 7) return date.toLocaleDateString(locale, { weekday: 'long' });
  return date.toLocaleDateString(locale, { day: 'numeric', month: 'short' });
}
