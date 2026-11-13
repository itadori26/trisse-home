import { parseEntries } from '../domain/frecency';
import type { UsageState, UsageStore } from '../usecase/searchUsage';

const KEY = 'trisse-home.search-usage';

/**
 * Keeps the search history in localStorage (#159). Works without storage
 * (private windows, blocked site data): then nothing is remembered.
 */
export class LocalUsageStore implements UsageStore {
  constructor(private readonly storage: () => Storage | undefined = () => globalThis.localStorage) {}

  load(): UsageState {
    try {
      const raw = this.storage()?.getItem(KEY);
      if (!raw) return { enabled: true, entries: [] };
      const data: unknown = JSON.parse(raw);
      const obj = typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : {};
      return { enabled: obj['enabled'] !== false, entries: parseEntries(obj['entries']) };
    } catch {
      return { enabled: true, entries: [] };
    }
  }

  save(state: UsageState): void {
    try {
      this.storage()?.setItem(KEY, JSON.stringify(state));
    } catch {
      // Storage full or blocked: the history simply is not kept.
    }
  }
}
