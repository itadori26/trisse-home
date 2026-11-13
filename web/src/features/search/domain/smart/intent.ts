import { fold, matchToken, QUALITY_FACTOR, tokenize } from './text';

/**
 * Intents of the smart search as data (#156): weighted terms in German and
 * English, typed slots (#158), the kind of answer and the widget that
 * shows the source. Scoring and ranking are pure functions.
 */

/** Kinds of answers. A new kind is added only when at least three sources need it. */
export type AnswerKind = 'metric' | 'status' | 'list' | 'text' | 'actions';

export type SlotType = 'entity' | 'place' | 'day' | 'number' | 'action';

export interface SlotSpec {
  name: string;
  type: SlotType;
  required?: boolean;
  /** Entity kind for entity slots, e.g. "container" or "service". */
  entityKind?: string;
}

export interface IntentSpec {
  id: string;
  /** Source that resolves the intent. */
  source: string;
  /** Weighted terms (folded, German and English). */
  terms: { term: string; weight: number }[];
  slots?: SlotSpec[];
  answer: AnswerKind;
  /** Widget kind that shows the details ("open widget"). */
  widgetKind?: string;
  /** Explicit prefixes that select the intent directly, e.g. "cpu" for "/cpu". */
  prefixes?: string[];
}

/** Something with a name the user can refer to (container, service …). */
export interface Entity {
  kind: string;
  id: string;
  name: string;
}

export type Action = 'start' | 'stop' | 'restart';

export type SlotValue =
  | { type: 'entity'; entity: Entity }
  | { type: 'place'; text: string }
  | { type: 'day'; offset: number }
  | { type: 'number'; value: number }
  | { type: 'action'; action: Action };

/** An interpreted query: intent, score and filled slots. */
export interface Match {
  intent: IntentSpec;
  score: number;
  slots: Record<string, SlotValue>;
  /** Required slots that could not be filled (asks back, #158). */
  missing: string[];
  /** Entities of the missing kind that come close, for suggestions. */
  suggestions: Entity[];
}

/** Minimum score of an answer and the number of answers shown. */
export const THRESHOLD = 1;
export const MAX_ANSWERS = 3;

const ACTION_WORDS: Record<string, Action> = {
  restart: 'restart',
  neustart: 'restart',
  neustarten: 'restart',
  reboot: 'restart',
  start: 'start',
  starte: 'start',
  starten: 'start',
  stop: 'stop',
  stopp: 'stop',
  stoppe: 'stop',
  stoppen: 'stop',
  anhalten: 'stop',
};

const DAY_WORDS: Record<string, number> = {
  heute: 0,
  today: 0,
  jetzt: 0,
  now: 0,
  morgen: 1,
  tomorrow: 1,
  ubermorgen: 2,
};

const WEEKDAYS: string[][] = [
  ['sonntag', 'sunday', 'so', 'sun'],
  ['montag', 'monday', 'mo', 'mon'],
  ['dienstag', 'tuesday', 'di', 'tue'],
  ['mittwoch', 'wednesday', 'mi', 'wed'],
  ['donnerstag', 'thursday', 'do', 'thu'],
  ['freitag', 'friday', 'fr', 'fri'],
  ['samstag', 'saturday', 'sa', 'sat'],
];

/** Days from today to a day word ("morgen", "freitag" …), or null. */
export function dayOffset(token: string, today: Date): number | null {
  const fixed = DAY_WORDS[token];
  if (fixed !== undefined) return fixed;
  // Only full weekday names; two-letter abbreviations collide with words.
  const wd = WEEKDAYS.findIndex((names) => names.slice(0, 2).includes(token));
  if (wd < 0) return null;
  return (wd - today.getDay() + 7) % 7;
}

/** Joins "neu starten" to one action word before tokens are matched. */
function joinPhrases(tokens: string[]): string[] {
  const out: string[] = [];
  for (let i = 0; i < tokens.length; i++) {
    if (tokens[i] === 'neu' && tokens[i + 1]?.startsWith('start')) {
      out.push('neustarten'); // the verb; "Neustart" alone is the noun (uptime)
      i++;
    } else out.push(tokens[i] ?? '');
  }
  return out;
}

interface EntityHit {
  entity: Entity;
  quality: number;
  used: number[];
}

/** Best entity of a kind among the unused tokens (names may have several words). */
function findEntity(
  tokens: string[],
  used: Set<number>,
  entities: readonly Entity[],
  kind: string,
): EntityHit | null {
  let best: EntityHit | null = null;
  for (const entity of entities) {
    if (entity.kind !== kind) continue;
    const name = fold(entity.name);
    const parts = tokenize(entity.name);
    for (let i = 0; i < tokens.length; i++) {
      if (used.has(i)) continue;
      const token = tokens[i] ?? '';
      // The whole name, or one of its words (e.g. "home" for "Home Assistant").
      const q = matchToken(token, name) ?? parts.map((p) => matchToken(token, p)).find((x) => x) ?? null;
      if (!q) continue;
      const quality = QUALITY_FACTOR[q] * (name === token ? 1.2 : 1);
      if (!best || quality > best.quality) best = { entity, quality, used: [i] };
    }
  }
  return best;
}

/** Interprets the query against one intent; null when nothing matched. */
function interpretIntent(
  tokens: string[],
  intent: IntentSpec,
  entities: readonly Entity[],
  today: Date,
  explicit: boolean,
): Match | null {
  const used = new Set<number>();
  let score = explicit ? 100 : 0;
  // Points from the intent's own words (terms, action words); a match that
  // lacks a required slot is only useful when the user asked for it.
  let trigger = score;
  if (!explicit) {
    for (const { term, weight } of intent.terms) {
      let best = 0;
      let at = -1;
      tokens.forEach((t, i) => {
        if (used.has(i)) return;
        const q = matchToken(t, term);
        const s = q ? weight * QUALITY_FACTOR[q] : 0;
        if (s > best) [best, at] = [s, i];
      });
      if (at >= 0) {
        used.add(at);
        score += best;
        trigger += best;
      }
    }
  }
  const slots: Record<string, SlotValue> = {};
  const missing: string[] = [];
  let suggestions: Entity[] = [];
  for (const slot of intent.slots ?? []) {
    const free = tokens.map((t, i) => ({ t, i })).filter(({ i }) => !used.has(i));
    let value: SlotValue | null = null;
    if (slot.type === 'entity' && slot.entityKind) {
      const hit = findEntity(tokens, used, entities, slot.entityKind);
      if (hit) {
        hit.used.forEach((i) => used.add(i));
        value = { type: 'entity', entity: hit.entity };
        // A known name counts strongly: "jellyfin" alone selects its status.
        score += 2 * hit.quality;
      } else suggestions = entities.filter((e) => e.kind === slot.entityKind).slice(0, 5);
    } else if (slot.type === 'action') {
      const f = free.find(({ t }) => ACTION_WORDS[t]);
      if (f) {
        used.add(f.i);
        const action = ACTION_WORDS[f.t] as Action;
        // "starte pihole neu" is a restart.
        value = { type: 'action', action: action === 'start' && tokens.includes('neu') ? 'restart' : action };
        score += 1;
        trigger += 1;
      }
    } else if (slot.type === 'day') {
      const f = free.find(({ t }) => dayOffset(t, today) !== null);
      if (f) {
        used.add(f.i);
        value = { type: 'day', offset: dayOffset(f.t, today) ?? 0 };
      }
    } else if (slot.type === 'number') {
      const f = free.find(({ t }) => /^\d+([.,]\d+)?$/.test(t));
      if (f) {
        used.add(f.i);
        value = { type: 'number', value: Number(f.t.replace(',', '.')) };
      }
    } else if (slot.type === 'place') {
      // Whatever is left after terms and the other slots, in the order typed.
      const rest = tokens.filter((_, i) => !used.has(i) && !ACTION_WORDS[tokens[i] ?? '']);
      const others = (intent.slots ?? []).filter((s) => s.type === 'day');
      const words = rest.filter((t) => !(others.length > 0 && dayOffset(t, today) !== null));
      if (words.length > 0) value = { type: 'place', text: words.join(' ') };
    }
    if (value) slots[slot.name] = value;
    else if (slot.required) missing.push(slot.name);
  }
  if (score <= 0 || (missing.length > 0 && trigger === 0)) return null;
  return { intent, score, slots, missing, suggestions: missing.length ? suggestions : [] };
}

/**
 * Interprets a query: an explicit prefix ("/cpu", "/docker jellyfin")
 * selects its intents directly (chain of responsibility); otherwise every
 * intent collects points for matching words and entity names.
 */
export function interpret(
  query: string,
  intents: readonly IntentSpec[],
  entities: readonly Entity[],
  today: Date,
): Match[] {
  const trimmed = query.trim();
  if (trimmed.startsWith('/')) {
    const [head = '', ...rest] = trimmed.slice(1).split(/\s+/);
    const prefix = fold(head);
    const restTokens = joinPhrases(tokenize(rest.join(' ')));
    const direct = intents.filter((i) => i.prefixes?.includes(prefix));
    if (direct.length > 0) {
      // The best interpretation of the prefix only, e.g. "/docker jellyfin"
      // is the status, "/docker restart jellyfin" the action.
      const best = direct
        .map((i) => interpretIntent(restTokens, i, entities, today, true))
        .filter((m): m is Match => m !== null)
        .map((m) => ({ ...m, score: m.score + Object.keys(m.slots).length - 2 * m.missing.length }))
        .sort((a, b) => b.score - a.score)[0];
      return best ? [best] : [];
    }
  }
  const tokens = joinPhrases(tokenize(trimmed));
  if (tokens.length === 0) return [];
  return intents
    .map((i) => interpretIntent(tokens, i, entities, today, false))
    .filter((m): m is Match => m !== null);
}

/** Keeps matches above the threshold, best first, one per intent, at most three. */
export function rank(matches: readonly Match[]): Match[] {
  return [...matches]
    .filter((m) => m.score >= THRESHOLD)
    .sort((a, b) => b.score - a.score || a.intent.id.localeCompare(b.intent.id))
    .slice(0, MAX_ANSWERS);
}
