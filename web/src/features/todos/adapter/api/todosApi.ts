import type { HttpClient } from '../../../../shared/api-client/http';
import type { TodoDTO } from '../../../../shared/api-client/generated/todos';
import type { ClearedDTO, TodoInputDTO } from '../../../../shared/api-client/generated/todos-requests';
import type { Reminder, Repeat, RepeatKind, Todo } from '../../domain/todo';
import type { TodosGateway } from '../../usecase/todosUseCases';

const toTodo = (d: TodoDTO): Todo => ({
  id: d.id,
  title: d.title,
  due: d.due,
  done: d.done,
  overdue: d.overdue,
  repeat: {
    kind: (d.repeat?.kind ?? '') as RepeatKind,
    every: d.repeat?.every || 1,
    days: d.repeat?.days ?? [],
  },
  reminder: { at: d.reminder?.at ?? '', daysBefore: d.reminder?.daysBefore ?? 0 },
  version: d.version ?? 0,
});

/** TodosGateway backed by the REST API. */
export class TodosApi implements TodosGateway {
  constructor(private readonly http: HttpClient) {}

  async list(includeDone: boolean): Promise<Todo[]> {
    return (await this.http.get<TodoDTO[]>(`/api/todos${includeDone ? '?done=true' : ''}`)).map(toTodo);
  }
  async create(title: string, due: string, id?: string): Promise<Todo> {
    // A client-generated id makes a replayed create idempotent (#80).
    const body: TodoInputDTO = { title, due, done: false, ...(id ? { id } : {}) };
    return toTodo(await this.http.post<TodoDTO>('/api/todos', body));
  }
  async update(
    id: string,
    title: string,
    due: string,
    done: boolean,
    schedule?: { repeat: Repeat; reminder: Reminder },
    version?: number,
  ): Promise<Todo> {
    // Without a schedule the server keeps the stored recurrence and reminder;
    // with a version it refuses changes made from an older one (409, #80).
    const body: TodoInputDTO = { title, due, done, ...schedule, ...(version ? { version } : {}) };
    return toTodo(await this.http.put<TodoDTO>(`/api/todos/${encodeURIComponent(id)}`, body));
  }
  delete(id: string): Promise<void> {
    return this.http.delete(`/api/todos/${encodeURIComponent(id)}`);
  }
  async clearCompleted(): Promise<number> {
    return (await this.http.post<ClearedDTO>('/api/todos/clear-completed')).removed;
  }
}
