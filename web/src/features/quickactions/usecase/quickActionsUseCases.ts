import { isoDate } from '../domain/dates';
import { parseQuick, type Prefixes, type QuickCommand } from '../domain/command';

/**
 * Output port: the features a quick command creates something in. Wired
 * in the composition root, so this feature imports none of them (#105).
 */
export interface QuickTargets {
  addTodo(title: string, due: string): Promise<{ id: string }>;
  deleteTodo(id: string): Promise<void>;
  addNote(title: string): Promise<{ id: string }>;
  deleteNote(id: string): Promise<void>;
  addBookmark(url: string, title: string): Promise<{ id: string }>;
  deleteBookmark(id: string): Promise<void>;
  /** Containers the restart action is allowed for (server allow-list). */
  restartableContainers(): Promise<string[]>;
  restartContainer(name: string): Promise<void>;
  /** Services with a MAC address that can be woken. */
  wakeableServices(): Promise<{ id: string; name: string }[]>;
  wakeService(id: string): Promise<void>;
}

/** Result of an executed command. */
export interface QuickResult {
  command: QuickCommand;
  /** Removes what was created; absent for actions. */
  undo?: () => Promise<void>;
  /** Link to open the created item (bookmarks). */
  openUrl?: string;
}

/** Thrown when an action target is not allowed or unknown. */
export class QuickActionError extends Error {}

/** Clock port. */
export interface Today {
  now(): Date;
}

/** Application logic of quick commands (#105, #67). */
export class QuickActionsUseCases {
  constructor(
    private readonly targets: QuickTargets,
    private readonly clock: Today = { now: () => new Date() },
  ) {}

  /** Parses input with the given prefixes; null means "search instead". */
  parse(input: string, prefixes: Prefixes): QuickCommand | null {
    return parseQuick(input, prefixes, this.clock.now());
  }

  /** Today as YYYY-MM-DD (for previews). */
  today(): string {
    return isoDate(this.clock.now());
  }

  /**
   * Executes a command. Restart and wake only work for allow-listed targets
   * (the server checks again); the caller must have asked for confirmation.
   */
  async execute(c: QuickCommand): Promise<QuickResult> {
    switch (c.kind) {
      case 'todo': {
        const { id } = await this.targets.addTodo(c.title, c.due);
        return { command: c, undo: () => this.targets.deleteTodo(id) };
      }
      case 'note': {
        const { id } = await this.targets.addNote(c.title);
        return { command: c, undo: () => this.targets.deleteNote(id) };
      }
      case 'bookmark': {
        const { id } = await this.targets.addBookmark(c.url, c.title);
        return { command: c, undo: () => this.targets.deleteBookmark(id), openUrl: c.url };
      }
      case 'restart': {
        const allowed = await this.targets.restartableContainers();
        const name = allowed.find((n) => n.toLowerCase() === c.target.toLowerCase());
        if (!name) throw new QuickActionError('restart');
        await this.targets.restartContainer(name);
        return { command: { ...c, target: name } };
      }
      case 'wake': {
        const services = await this.targets.wakeableServices();
        const s = services.find((x) => x.name.toLowerCase() === c.target.toLowerCase());
        if (!s) throw new QuickActionError('wake');
        await this.targets.wakeService(s.id);
        return { command: { ...c, target: s.name } };
      }
    }
  }
}
