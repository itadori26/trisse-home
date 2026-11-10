import type { HttpClient } from '../../../../shared/api-client/http';
import type { TodoDTO } from '../../../../shared/api-client/generated/todos';
import type { ClearedDTO, TodoInputDTO } from '../../../../shared/api-client/generated/todos-requests';
import type { Todo } from '../../domain/todo';
import type { TodosGateway } from '../../usecase/todosUseCases';

const toTodo = (d: TodoDTO): Todo => ({
  id: d.id,
  title: d.title,
  due: d.due,
  done: d.done,
  overdue: d.overdue,
});

/** TodosGateway backed by the REST API. */
export class TodosApi implements TodosGateway {
  constructor(private readonly http: HttpClient) {}

  async list(includeDone: boolean): Promise<Todo[]> {
    return (await this.http.get<TodoDTO[]>(`/api/todos${includeDone ? '?done=true' : ''}`)).map(toTodo);
  }
  async create(title: string, due: string): Promise<Todo> {
    const body: TodoInputDTO = { title, due, done: false };
    return toTodo(await this.http.post<TodoDTO>('/api/todos', body));
  }
  async update(id: string, title: string, due: string, done: boolean): Promise<Todo> {
    const body: TodoInputDTO = { title, due, done };
    return toTodo(await this.http.put<TodoDTO>(`/api/todos/${encodeURIComponent(id)}`, body));
  }
  delete(id: string): Promise<void> {
    return this.http.delete(`/api/todos/${encodeURIComponent(id)}`);
  }
  async clearCompleted(): Promise<number> {
    return (await this.http.post<ClearedDTO>('/api/todos/clear-completed')).removed;
  }
}
