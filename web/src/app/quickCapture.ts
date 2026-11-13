import { formatDue } from '../features/todos/domain/todo';
import type { QuickActionsUseCases, QuickCommand } from '../features/quickactions';
import { intlLocale, t } from '../shared/i18n';

/** What a quick command does, in words, e.g. for previews and toasts (#105, #67). */
export function describeCommand(c: QuickCommand, uc: QuickActionsUseCases, done = false): string {
  switch (c.kind) {
    case 'todo': {
      const due = c.due ? ` (${formatDue(c.due, new Date(`${uc.today()}T12:00:00`), intlLocale())})` : '';
      return t(done ? 'quick.done_todo' : 'quick.todo', { title: c.title }) + due;
    }
    case 'note':
      return t(done ? 'quick.done_note' : 'quick.note', { title: c.title });
    case 'bookmark':
      return t(done ? 'quick.done_bookmark' : 'quick.bookmark', { title: c.title, url: c.url });
    case 'restart':
      return t(done ? 'quick.done_restart' : 'quick.restart', { name: c.target });
    case 'wake':
      return t(done ? 'quick.done_wake' : 'quick.wake', { name: c.target });
  }
}
