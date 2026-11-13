import type { LauncherSources } from '../features/launcher/usecase/launcherUseCases';
import type { Container } from './container';

/**
 * Adapters of the launcher ports (#189) on top of the existing services and
 * bookmarks use cases – composition root, so no feature imports another.
 * Only data the dashboard already loads; no new outbound requests.
 */
export function launcherSources(c: Container, modules: Set<string>): LauncherSources {
  return {
    services: modules.has('services')
      ? async () =>
          (await c.services.list()).map((s) => ({
            id: s.id,
            name: s.name,
            url: s.url,
            state: s.state,
            latencyMs: s.latencyMs,
          }))
      : null,
    bookmarks: modules.has('bookmarks')
      ? async () => {
          const l = await c.bookmarks.list();
          const folders = new Map(l.folders.map((f) => [f.id, f.name]));
          return l.bookmarks.map((b) => ({
            id: b.id,
            title: b.title,
            url: b.url,
            folder: folders.get(b.folderId) ?? '',
          }));
        }
      : null,
  };
}
