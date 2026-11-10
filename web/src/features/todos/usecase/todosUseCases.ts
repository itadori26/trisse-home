import type { Todo } from '../domain/todo';

/** Output port: to-do backend. */
export interface TodosGateway {
  list(includeDone: boolean): Promise<Todo[]>;
  create(title: string, due: string): Promise<Todo>;
  update(id: string, title: string, due: string, done: boolean): Promise<Todo>;
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

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  clearCompleted(): Promise<number> {
    return this.gateway.clearCompleted();
  }
}
