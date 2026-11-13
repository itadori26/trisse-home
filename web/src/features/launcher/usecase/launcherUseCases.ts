import type { LauncherBookmark, LauncherService } from '../domain/launcher';

/**
 * Output ports of the launcher (#189), implemented in the composition root
 * on top of the services and bookmarks use cases; a disabled module is
 * simply absent (null).
 */
export interface LauncherSources {
  services: (() => Promise<LauncherService[]>) | null;
  bookmarks: (() => Promise<LauncherBookmark[]>) | null;
}

export interface LauncherData {
  services: LauncherService[];
  bookmarks: LauncherBookmark[];
}

/** Application logic of the launcher. */
export class LauncherUseCases {
  constructor(private readonly sources: LauncherSources) {}

  /** Whether the services module is enabled (status pill). */
  get hasServices(): boolean {
    return this.sources.services !== null;
  }

  /** Loads both lists; a failing source just stays empty. */
  async load(): Promise<LauncherData> {
    const [services, bookmarks] = await Promise.all([
      this.sources.services?.().catch(() => []) ?? [],
      this.sources.bookmarks?.().catch(() => []) ?? [],
    ]);
    return { services, bookmarks };
  }

  /** Services only, for the status pill of the top bar. */
  async services(): Promise<LauncherService[]> {
    return this.sources.services ? this.sources.services() : [];
  }
}
