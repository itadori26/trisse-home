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
  label: string;
  items: QuickItem[];
}

const order: QuickKind[] = ['bookmark', 'service', 'note', 'todo'];
const labels: Record<QuickKind, string> = {
  bookmark: 'Bookmarks',
  service: 'Services',
  note: 'Notes',
  todo: 'To-dos',
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

/** Ranks the items, groups them by kind and limits each group. */
export function search(items: readonly QuickItem[], query: string, perGroup = 5): QuickGroup[] {
  const scored = items
    .map((item) => ({ item, s: score(item, query) }))
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

/** Whether a keyboard event should open the quick search. */
export function opensQuickSearch(
  e: Pick<KeyboardEvent, 'key' | 'ctrlKey' | 'metaKey' | 'altKey'>,
  target: EventTarget | null,
): boolean {
  if ((e.ctrlKey || e.metaKey) && !e.altKey && e.key.toLowerCase() === 'k') return true;
  if (e.key !== '/' || e.ctrlKey || e.metaKey || e.altKey) return false;
  const el = target as HTMLElement | null;
  const editable =
    !!el && (el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName ?? ''));
  return !editable;
}
