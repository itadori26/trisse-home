import { type Reminder, type Repeat, type Todo, validateSchedule } from '../domain/todo';

/** Output port: to-do backend. */
export interface TodosGateway {
  list(includeDone: boolean): Promise<Todo[]>;
  create(title: string, due: string): Promise<Todo>;
  update(
    id: string,
    title: string,
    due: string,
    done: boolean,
    schedule?: { repeat: Repeat; reminder: Reminder },
  ): Promise<Todo>;
  delete(id: string): Promise<void>;
  clearCompleted(): Promise<number>;
}

/** Application logic of the to-do feature. */
export class TodosUseCases {
  constructor(private readonly gateway: TodosGateway) {}

  list(includeDone: boolean): Promise<Todo[]> {
    return this.gateway.list(includeDone);
  }

  /**
   * Adds a to-do. A trailing "@YYYY-MM-DD" in the text sets the due date,
   * e.g. "Pay rent @2026-12-01".
   */
  add(text: string): Promise<Todo> {
    const match = /\s@(\d{4}-\d{2}-\d{2})\s*$/.exec(text);
    const title = match ? text.slice(0, match.index) : text;
    return this.gateway.create(title.trim(), match?.[1] ?? '');
  }

  toggle(todo: Todo): Promise<Todo> {
    return this.gateway.update(todo.id, todo.title, todo.due, !todo.done);
  }

  edit(todo: Todo, title: string, due: string): Promise<Todo> {
    return this.gateway.update(todo.id, title.trim(), due, todo.done);
  }

  /** Changes text, due date, recurrence and reminder (#68). */
  async editFull(
    todo: Todo,
    change: { title: string; due: string; repeat: Repeat; reminder: Reminder },
  ): Promise<Todo> {
    const repeat: Repeat =
      change.repeat.kind === ''
        ? { kind: '', every: 0, days: [] }
        : {
            ...change.repeat,
            days: change.repeat.kind === 'weekly' ? [...new Set(change.repeat.days)].sort() : [],
          };
    const problem = validateSchedule(change.due, change.repeat, change.reminder);
    if (problem) throw new Error(problem);
    return this.gateway.update(todo.id, change.title.trim(), change.due, todo.done, {
      repeat,
      reminder: change.reminder,
    });
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  clearCompleted(): Promise<number> {
    return this.gateway.clearCompleted();
  }
}
