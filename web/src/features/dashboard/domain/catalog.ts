import type { WidgetType } from './dashboard';

/** Categories of the widget picker in their display order (#104). */
export const widgetCategories = ['everyday', 'information', 'server', 'network'] as const;
export type WidgetCategory = (typeof widgetCategories)[number];

/** Widget types of one category. */
export interface CategoryGroup {
  category: WidgetCategory;
  types: WidgetType[];
}

/** Groups types by category in the fixed order, keeping their order; empty categories are omitted. */
export function groupByCategory(types: readonly WidgetType[]): CategoryGroup[] {
  return widgetCategories
    .map((category) => ({ category, types: types.filter((t) => t.category === category) }))
    .filter((g) => g.types.length > 0);
}

/** Lower case without accents, so "Zeitplan" matches "zeitplan" and "é" matches "e". */
export function normalise(s: string): string {
  return s
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase()
    .trim();
}

/**
 * Filters types by a query over title, keywords and description. Every
 * word must match somewhere; results are ranked title before keywords
 * before description, ties keep the original order.
 */
export function searchWidgetTypes(types: readonly WidgetType[], query: string): WidgetType[] {
  const words = normalise(query).split(/\s+/).filter(Boolean);
  if (words.length === 0) return [...types];
  const scored: { t: WidgetType; score: number; i: number }[] = [];
  types.forEach((t, i) => {
    const title = normalise(t.title);
    const keywords = t.keywords.map(normalise);
    const description = normalise(t.description);
    let score = 0;
    for (const w of words) {
      if (title.includes(w)) score += title.startsWith(w) ? 0 : 1;
      else if (keywords.some((k) => k.includes(w))) score += 10;
      else if (description.includes(w)) score += 100;
      else return; // a word without any match excludes the type
    }
    scored.push({ t, score, i });
  });
  return scored.sort((a, b) => a.score - b.score || a.i - b.i).map((s) => s.t);
}

/**
 * Keyboard navigation in the picker grid: columns of tiles, arrow keys move
 * within a column and across columns (keeping the row where possible).
 * Returns the [column, row] to focus next.
 */
export function moveFocus(
  columns: readonly number[],
  at: [number, number],
  key: 'ArrowUp' | 'ArrowDown' | 'ArrowLeft' | 'ArrowRight' | 'Home' | 'End',
): [number, number] {
  const [c, r] = at;
  const len = (i: number) => columns[i] ?? 0;
  switch (key) {
    case 'ArrowDown':
      if (r + 1 < len(c)) return [c, r + 1];
      return c + 1 < columns.length ? [c + 1, 0] : [c, r];
    case 'ArrowUp':
      if (r > 0) return [c, r - 1];
      return c > 0 ? [c - 1, len(c - 1) - 1] : [c, r];
    case 'ArrowRight':
      return c + 1 < columns.length ? [c + 1, Math.min(r, len(c + 1) - 1)] : [c, r];
    case 'ArrowLeft':
      return c > 0 ? [c - 1, Math.min(r, len(c - 1) - 1)] : [c, r];
    case 'Home':
      return [0, 0];
    case 'End':
      return [columns.length - 1, len(columns.length - 1) - 1];
  }
}
