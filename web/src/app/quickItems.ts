import { t } from '../shared/i18n';
import type { Container } from './container';
import type { QuickItem } from './quickSearch';

type Todo = Parameters<Container['todos']['toggle']>[0];

/** What one loader adds to the quick search. */
export interface QuickItemsResult {
  items: QuickItem[];
  /** Full notes for the preview. */
  notes?: Map<string, { title: string; body: string }>;
  /** Open to-dos so a result can be ticked off. */
  todos?: Map<string, Todo>;
}

/**
 * Loaders of the classic quick search results, one per enabled module.
 * They live in the composition root so the dialog itself calls no module
 * (#156); each loader fails on its own.
 */
export function quickItemLoaders(c: Container, modules: Set<string>): Promise<QuickItemsResult>[] {
  const loaders: Promise<QuickItemsResult>[] = [];
  if (modules.has('bookmarks'))
    loaders.push(
      c.bookmarks.list().then((l) => {
        const folders = new Map(l.folders.map((f) => [f.id, f.name]));
        return {
          items: l.bookmarks.map((b) => ({
            kind: 'bookmark' as const,
            id: b.id,
            title: b.title,
            detail: [folders.get(b.folderId), b.url].filter(Boolean).join(' · '),
            keywords: [...b.tags, b.description],
            url: b.url,
          })),
        };
      }),
    );
  if (modules.has('services'))
    loaders.push(
      c.services.list().then((l) => ({
        items: l.map((s) => ({
          kind: 'service' as const,
          id: s.id,
          title: s.name,
          detail: s.url,
          keywords: [s.description],
          url: s.url,
        })),
      })),
    );
  if (modules.has('notes'))
    loaders.push(
      c.notes.list('all').then((l) => ({
        notes: new Map(l.map((n) => [n.id, { title: n.title, body: n.body }])),
        items: l.map((n) => ({
          kind: 'note' as const,
          id: n.id,
          title: n.title || t('app.untitled_note'),
          detail: n.body.slice(0, 200),
        })),
      })),
    );
  if (modules.has('todos'))
    loaders.push(
      c.todos.list(false).then((l) => ({
        todos: new Map(l.map((x) => [x.id, x])),
        items: l.map((x) => ({
          kind: 'todo' as const,
          id: x.id,
          title: x.title,
          detail: x.due ? `due ${x.due}` : '',
        })),
      })),
    );
  return loaders;
}
