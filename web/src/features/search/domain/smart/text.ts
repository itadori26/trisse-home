/**
 * Text building blocks of the smart search (#156, #157): normalising,
 * tokenising and tolerant matching of words. Pure functions only.
 */

/** Lower-cased text without diacritics, so "Büro" matches "buro". */
export function fold(s: string): string {
  return s
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase()
    .replace(/ß/g, 'ss');
}

/** Words that carry no meaning for an intent (German and English). */
const STOP_WORDS = new Set(
  (
    'der die das den dem des ein eine einen einem einer ist sind wie was wo wann wer welche welcher welches ' +
    'mein meine meinen meiner mir mich ich du es im in am an auf zu zum zur mit von vom fur bei noch gerade ' +
    'jetzt aktuell bitte mal denn doch ja und oder ' +
    'the a an is are how what where when who which my me i you it in on at to of for with please now current ' +
    'currently show tell and or'
  ).split(' '),
);

/** Splits a query into folded words without punctuation and stop words. */
export function tokenize(query: string): string[] {
  return fold(query)
    .split(/[^\p{L}\p{N}._-]+/u)
    .map((t) => t.replace(/^[._-]+|[._-]+$/g, ''))
    .filter((t) => t !== '' && !STOP_WORDS.has(t));
}

/** How a token matched a term; ordered from best to weakest. */
export type MatchQuality = 'exact' | 'prefix' | 'fuzzy';

/** Points per quality, so exact > prefix > fuzzy in the score. */
export const QUALITY_FACTOR: Record<MatchQuality, number> = { exact: 1, prefix: 0.8, fuzzy: 0.6 };

/** Minimum token length for prefix matches. */
export const MIN_PREFIX = 2;

/** Allowed edit distance for a term length: ≤ 1 from 4, ≤ 2 from 8 characters. */
export function maxDistance(length: number): number {
  if (length >= 8) return 2;
  if (length >= 4) return 1;
  return 0;
}

/**
 * Damerau-Levenshtein distance (optimal string alignment: insertion,
 * deletion, substitution, transposition of neighbours). Returns
 * limit + 1 as soon as the distance must exceed the limit, so obvious
 * mismatches never fill the whole matrix.
 */
export function boundedDistance(a: string, b: string, limit: number): number {
  if (Math.abs(a.length - b.length) > limit) return limit + 1;
  let prev2: number[] = [];
  let prev = Array.from({ length: b.length + 1 }, (_, j) => j);
  for (let i = 1; i <= a.length; i++) {
    const cur = [i];
    let rowMin = i;
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      let v = Math.min((prev[j] ?? 0) + 1, (cur[j - 1] ?? 0) + 1, (prev[j - 1] ?? 0) + cost);
      if (i > 1 && j > 1 && a[i - 1] === b[j - 2] && a[i - 2] === b[j - 1]) {
        v = Math.min(v, (prev2[j - 2] ?? 0) + 1);
      }
      cur[j] = v;
      rowMin = Math.min(rowMin, v);
    }
    if (rowMin > limit) return limit + 1;
    prev2 = prev;
    prev = cur;
  }
  return prev[b.length] ?? limit + 1;
}

/**
 * Matches a (folded) token against a (folded) term: exact, as a prefix
 * while typing (from two characters), or with a few typos. Short tokens
 * (≤ 3 characters) only match exactly or by prefix to avoid noise.
 */
export function matchToken(token: string, term: string): MatchQuality | null {
  if (token === term) return 'exact';
  if (token.length >= MIN_PREFIX && term.startsWith(token)) return 'prefix';
  if (token.length <= 3) return null;
  const limit = maxDistance(term.length);
  if (limit === 0) return null;
  return boundedDistance(token, term, limit) <= limit ? 'fuzzy' : null;
}
