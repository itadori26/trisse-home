import type { ActionItem, Answer } from '../domain/smart/answer';
import {
  type Entity,
  type IntentSpec,
  interpret,
  MAX_ANSWERS,
  type Match,
  rank,
  THRESHOLD,
} from '../domain/smart/intent';

/**
 * Port of one answer source (Strategy, #156), implemented in the
 * composition root on top of the feature use cases. Disabled modules
 * register no source.
 */
export interface SearchSource {
  id: string;
  /** Names the user can refer to; loaded once per session. */
  entities?(): Promise<Entity[]>;
  /** Builds the answer for a match; must honour the abort signal. */
  resolve?(match: Match, signal: AbortSignal): Promise<Answer>;
  /**
   * Answers the raw query directly and synchronously, for sources that do
   * not fit the intent model, e.g. the calculator (#160). Null means "not
   * mine"; direct answers come before the intent answers.
   */
  direct?(query: string): Answer | null;
  /** Runs a confirmed write action. */
  execute?(item: ActionItem): Promise<void>;
  /** How long a resolved answer is reused (default 15 s). */
  cacheMs?: number;
}

/** Builds the answer when a required slot is missing (#158). */
export type AskBack = (match: Match) => Answer;

export interface SmartSearchOptions {
  intents: readonly IntentSpec[];
  sources: readonly SearchSource[];
  /** Guests and kiosk mode get no write actions. */
  canAct: boolean;
  askBack: AskBack;
  now?: () => Date;
  /** Applies the learned boost to a match score (#159); target "answer:<intent>". */
  score?: (query: string, target: string, score: number) => number;
}

const DEFAULT_CACHE_MS = 15_000;

/** The smart search pipeline: interpret → rank → resolve (#156). */
export class SmartSearch {
  private entitiesLoad: Promise<Entity[]> | null = null;
  private entityList: Entity[] = [];
  private readonly cache = new Map<string, { at: number; answer: Promise<Answer> }>();
  private readonly bySource: Map<string, SearchSource>;
  private readonly intents: IntentSpec[];
  private readonly now: () => Date;

  constructor(private readonly opts: SmartSearchOptions) {
    this.bySource = new Map(opts.sources.map((s) => [s.id, s]));
    // Intents of disabled modules disappear with their source.
    this.intents = opts.intents.filter(
      (i) => this.bySource.has(i.source) && (opts.canAct || i.answer !== 'actions'),
    );
    this.now = opts.now ?? (() => new Date());
  }

  /** Loads the entity names once (when the search is first opened). */
  prepare(): Promise<Entity[]> {
    this.entitiesLoad ??= Promise.all(
      this.opts.sources.map((s) => (s.entities ? s.entities().catch(() => []) : Promise.resolve([]))),
    ).then((lists) => (this.entityList = lists.flat()));
    return this.entitiesLoad;
  }

  /** The ranked matches of a query, without resolving them (synchronous). */
  matches(query: string): Match[] {
    const found = interpret(query, this.intents, this.entityList, this.now());
    const score = this.opts.score;
    if (!score) return rank(found);
    // The boost reorders answers above the threshold; it never lifts one over it.
    return rank(
      found
        .filter((m) => m.score >= THRESHOLD)
        .map((m) => ({ ...m, score: score(query, `answer:${m.intent.id}`, m.score) })),
    );
  }

  /**
   * Answers a query. Sources run in parallel; one failing source drops its
   * answer only. A newer query aborts this one (the caller passes a fresh
   * signal per query, latest wins).
   */
  async answer(query: string, signal: AbortSignal): Promise<Answer[]> {
    const direct = this.direct(query);
    await this.prepare();
    signal.throwIfAborted();
    const matches = this.matches(query);
    const settled = await Promise.allSettled(matches.map((m) => this.resolve(m, signal)));
    if (signal.aborted) throw signal.reason ?? new DOMException('aborted', 'AbortError');
    const resolved = settled.flatMap((r) => (r.status === 'fulfilled' ? [r.value] : []));
    return [...direct, ...resolved].slice(0, MAX_ANSWERS);
  }

  /** Answers of the sources that read the raw query (synchronous, #160). */
  direct(query: string): Answer[] {
    const out: Answer[] = [];
    for (const s of this.opts.sources) {
      try {
        const a = s.direct?.(query);
        if (a) out.push(a);
      } catch {
        // A broken direct source must not break the search.
      }
    }
    return out;
  }

  private resolve(match: Match, signal: AbortSignal): Promise<Answer> {
    if (match.missing.length > 0) return Promise.resolve(this.opts.askBack(match));
    const source = this.bySource.get(match.intent.source);
    if (!source?.resolve) return Promise.reject(new Error(`no source ${match.intent.source}`));
    const key = `${match.intent.id}|${JSON.stringify(match.slots)}`;
    const ttl = source.cacheMs ?? DEFAULT_CACHE_MS;
    const now = this.now().getTime();
    const hit = this.cache.get(key);
    if (hit && now - hit.at < ttl) return hit.answer;
    const answer = source.resolve(match, signal);
    const entry = { at: now, answer };
    this.cache.set(key, entry);
    // Failures and aborted requests are not shared with the next query.
    const forget = () => this.cache.get(key) === entry && this.cache.delete(key);
    answer.catch(forget);
    signal.addEventListener('abort', forget, { once: true });
    return answer;
  }

  /** Runs a confirmed action and forgets the cached answers of its source. */
  async execute(item: ActionItem): Promise<void> {
    if (!this.opts.canAct) throw new Error('actions are not allowed');
    const source = this.bySource.get(item.source);
    if (!source?.execute) throw new Error(`source ${item.source} has no actions`);
    await source.execute(item);
    const prefixes = this.intents.filter((i) => i.source === item.source).map((i) => `${i.id}|`);
    for (const key of [...this.cache.keys()])
      if (prefixes.some((p) => key.startsWith(p))) this.cache.delete(key);
  }
}
