import type { BookmarksUseCases } from '../features/bookmarks/usecase/bookmarksUseCases';
import type { DockerUseCases } from '../features/docker/usecase/dockerUseCases';
import type { NotesUseCases } from '../features/notes/usecase/notesUseCases';
import type { QuickTargets } from '../features/quickactions/usecase/quickActionsUseCases';
import type { ServicesUseCases } from '../features/services/usecase/servicesUseCases';
import type { TodosUseCases } from '../features/todos/usecase/todosUseCases';

/** The features a quick command can reach (composition root, #105). */
interface Features {
  todos: TodosUseCases;
  notes: NotesUseCases;
  bookmarks: BookmarksUseCases;
  docker: DockerUseCases;
  services: ServicesUseCases;
}

/** Adapts the feature use cases to the quick actions port. */
export function quickTargets(f: Features): QuickTargets {
  return {
    addTodo: (title, due) => f.todos.add(due ? `${title} @${due}` : title),
    deleteTodo: (id) => f.todos.delete(id),
    addNote: (title) => f.notes.save({ title, body: '', pinned: false }),
    deleteNote: (id) => f.notes.delete(id),
    addBookmark: (url, title) => f.bookmarks.save({ folderId: '', title, url, description: '', tags: [] }),
    deleteBookmark: (id) => f.bookmarks.delete(id),
    restartableContainers: async () => {
      const a = await f.docker.actions();
      return a.enabled ? a.containers : [];
    },
    restartContainer: (name) => f.docker.restart(name),
    wakeableServices: async () =>
      (await f.services.list()).filter((s) => s.mac !== '').map((s) => ({ id: s.id, name: s.name })),
    wakeService: (id) => f.services.wake(id),
  };
}
