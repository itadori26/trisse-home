import { extractDue } from './dates';

/** A parsed quick command (#105, #67). */
export type QuickCommand =
  | { kind: 'todo'; title: string; due: string }
  | { kind: 'note'; title: string }
  | { kind: 'bookmark'; url: string; title: string }
  | { kind: 'restart'; target: string }
  | { kind: 'wake'; target: string };

export type QuickKind = QuickCommand['kind'];

/** Prefixes per command; matched case-insensitively. */
export type Prefixes = Partial<Record<QuickKind, string[]>>;

/**
 * Search bar prefixes: a letter followed by a space. Letters never collide
 * with DuckDuckGo !bangs, which start with "!".
 */
export const searchPrefixes: Prefixes = { todo: ['t'], note: ['n'], bookmark: ['b'] };

/** Command palette prefixes (#67). */
export const palettePrefixes: Prefixes = {
  todo: ['todo:', 't:'],
  note: ['note:', 'n:'],
  bookmark: ['bm:', 'b:'],
  restart: ['restart'],
  wake: ['wake'],
};

/** Reads configured search prefixes, e.g. { todo: "a", note: "n" }, falling back to the defaults. */
export function prefixesFromConfig(raw: unknown): Prefixes {
  const out: Prefixes = { ...searchPrefixes };
  if (raw && typeof raw === 'object') {
    for (const k of ['todo', 'note', 'bookmark'] as const) {
      const v = (raw as Record<string, unknown>)[k];
      if (typeof v === 'string' && /^[\p{L}\p{N}]{1,4}$/u.test(v.trim())) out[k] = [v.trim()];
    }
  }
  return out;
}

/**
 * Combines prefix sets, e.g. the palette's "todo:" with the search field's
 * short "t" (#221): the search field opens the palette, so both apply there.
 */
export function mergePrefixes(...sets: Prefixes[]): Prefixes {
  const out: Prefixes = {};
  for (const set of sets) {
    for (const [kind, list] of Object.entries(set) as [QuickKind, string[]][]) {
      out[kind] = [...new Set([...(out[kind] ?? []), ...list])];
    }
  }
  return out;
}

/** Turns "example.org/x" into "https://example.org/x"; null if it is no URL. */
export function normaliseUrl(raw: string): string | null {
  const s = raw.trim();
  if (!s || /\s/.test(s)) return null;
  const withScheme = /^[a-z][a-z0-9+.-]*:\/\//i.test(s) ? s : `https://${s}`;
  try {
    const u = new URL(withScheme);
    if ((u.protocol !== 'http:' && u.protocol !== 'https:') || !u.hostname.includes('.')) return null;
    return u.toString();
  } catch {
    return null;
  }
}

/**
 * Parses a quick command. A prefix only counts when followed by a space
 * (or directly by text for prefixes ending in ":"), so "tomorrow" never
 * becomes a to-do. Returns null when the input is a normal search.
 */
export function parseQuick(input: string, prefixes: Prefixes, today: Date): QuickCommand | null {
  const trimmed = input.trimStart();
  for (const [kind, list] of Object.entries(prefixes) as [QuickKind, string[]][]) {
    for (const p of list) {
      const lower = trimmed.toLowerCase();
      const pl = p.toLowerCase();
      if (!lower.startsWith(pl)) continue;
      const after = trimmed.slice(p.length);
      if (!p.endsWith(':') && !/^\s/.test(after)) continue;
      const rest = after.trim();
      if (!rest) continue;
      const cmd = build(kind, rest, today);
      if (cmd) return cmd;
    }
  }
  return null;
}

function build(kind: QuickKind, rest: string, today: Date): QuickCommand | null {
  switch (kind) {
    case 'todo': {
      const { text, due } = extractDue(rest, today);
      return text ? { kind, title: text.slice(0, 200), due } : null;
    }
    case 'note':
      return { kind, title: rest.slice(0, 200) };
    case 'bookmark': {
      const [first, ...titleWords] = rest.split(/\s+/);
      const url = normaliseUrl(first ?? '');
      if (!url) return null;
      return { kind, url, title: titleWords.join(' ') || new URL(url).hostname };
    }
    case 'restart':
    case 'wake':
      return { kind, target: rest };
  }
}
