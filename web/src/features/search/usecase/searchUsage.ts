import { boost, boosted, type FrecencyEntry, record } from '../domain/frecency';

/** What the usage store keeps: the switch and the recorded choices. */
export interface UsageState {
  enabled: boolean;
  entries: FrecencyEntry[];
}

/** Output port: where the search history lives (browser today, backend later). */
export interface UsageStore {
  load(): UsageState;
  save(state: UsageState): void;
}

/**
 * Learns from search choices (#159): records which result was chosen for
 * a query and boosts it the next time. Only choices are recorded – never
 * a query without a chosen result. Switched off, it neither records nor
 * boosts.
 */
export class SearchUsage {
  private state: UsageState;

  constructor(
    private readonly store: UsageStore,
    private readonly now: () => number = Date.now,
  ) {
    this.state = store.load();
  }

  get enabled(): boolean {
    return this.state.enabled;
  }

  /** Number of remembered choices (for the settings). */
  get size(): number {
    return this.state.entries.length;
  }

  /** The boost of a result for the query; 0 when switched off. */
  boost(query: string, target: string): number {
    return this.state.enabled ? boost(this.state.entries, query, target, this.now()) : 0;
  }

  /** A match score with the boost of its result applied. */
  score(query: string, target: string, score: number): number {
    return boosted(score, this.boost(query, target));
  }

  /** Remembers that `target` was chosen for `query`. */
  record(query: string, target: string): void {
    if (!this.state.enabled) return;
    this.update({ ...this.state, entries: record(this.state.entries, query, target, this.now()) });
  }

  setEnabled(enabled: boolean): void {
    // Switching off also forgets, so nothing lingers in the browser.
    this.update({ enabled, entries: enabled ? this.state.entries : [] });
  }

  clear(): void {
    this.update({ ...this.state, entries: [] });
  }

  private update(next: UsageState) {
    this.state = next;
    this.store.save(next);
  }
}
