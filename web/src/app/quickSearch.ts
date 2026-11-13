import type { MessageKey } from '../shared/i18n';

/** Kinds of entries the quick search finds. */
export type QuickKind = 'bookmark' | 'service' | 'note' | 'todo';

/** One searchable entry. */
export interface QuickItem {
  kind: QuickKind;
  id: string;
  title: string;
  /** Secondary text (URL, folder, due date …), also searched. */
  detail: string;
  /** Additional search terms, e.g. bookmark tags. */
  keywords?: string[];
  /** Target for bookmarks and services. */
  url?: string;
}

/** A group of results of one kind. */
export interface QuickGroup {
  kind: QuickKind;
  label: MessageKey;
  items: QuickItem[];
}

/** Id of an item in the search history (#159). */
export const itemTarget = (item: QuickItem) => `${item.kind}:${item.id}`;

const order: QuickKind[] = ['bookmark', 'service', 'note', 'todo'];
const labels: Record<QuickKind, MessageKey> = {
  bookmark: 'app.group.bookmarks',
  service: 'app.group.services',
  note: 'app.group.notes',
  todo: 'app.group.todos',
};

/** Lower-cased text without diacritics, so "Büro" matches "buro". */
export function fold(s: string): string {
  return s
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase();
}

/**
 * Scores an item for the query: 3 = title starts with it, 2 = a word of the
 * title starts with it, 1 = contained anywhere (title, detail, keywords),
 * 0 = no match. Every word of the query must match.
 */
export function score(item: QuickItem, query: string): number {
  const words = fold(query).split(/\s+/).filter(Boolean);
  if (words.length === 0) return 0;
  const title = fold(item.title);
  const haystack = [title, fold(item.detail), ...(item.keywords ?? []).map(fold)].join(' ');
  if (!words.every((w) => haystack.includes(w))) return 0;
  const first = words[0] ?? '';
  if (title.startsWith(first)) return 3;
  if (title.split(/[\s\-_/.]+/).some((t) => t.startsWith(first))) return 2;
  return 1;
}

/**
 * Ranks the items, groups them by kind and limits each group. `boost`
 * applies the learned frecency to a match score (#159).
 */
export function search(
  items: readonly QuickItem[],
  query: string,
  perGroup = 5,
  boost?: (item: QuickItem, score: number) => number,
): QuickGroup[] {
  const scored = items
    .map((item) => {
      const s = score(item, query);
      return { item, s: s > 0 && boost ? boost(item, s) : s };
    })
    .filter((x) => x.s > 0)
    .sort((a, b) => b.s - a.s || a.item.title.localeCompare(b.item.title));
  return order
    .map((kind) => ({
      kind,
      label: labels[kind],
      items: scored
        .filter((x) => x.item.kind === kind)
        .slice(0, perGroup)
        .map((x) => x.item),
    }))
    .filter((g) => g.items.length > 0);
}
