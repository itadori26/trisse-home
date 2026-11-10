/**
 * Pure layout algorithm of the dashboard grid (see ADR 0007). Items live on
 * a 12-column grid with unbounded rows; gravity pulls every item up as far
 * as possible without overlapping others.
 */

/** Number of grid columns; matches the backend. */
export const GRID_COLUMNS = 12;
/** Maximum height of a widget in rows; matches the backend. */
export const MAX_ROWS = 24;

/** Position and size in grid cells. */
export interface Rect {
  x: number;
  y: number;
  w: number;
  h: number;
}

/** A rectangle that belongs to a widget. */
export interface LayoutItem extends Rect {
  id: string;
}

const clamp = (v: number, min: number, max: number) => Math.min(Math.max(v, min), max);

/** Reports whether two rectangles share at least one cell. */
export function collides(a: Rect, b: Rect): boolean {
  return a.x < b.x + b.w && b.x < a.x + a.w && a.y < b.y + b.h && b.y < a.y + a.h;
}

/** Forces a rectangle into the grid bounds. */
export function clampRect(r: Rect): Rect {
  const w = clamp(Math.round(r.w), 1, GRID_COLUMNS);
  const h = clamp(Math.round(r.h), 1, MAX_ROWS);
  return { x: clamp(Math.round(r.x), 0, GRID_COLUMNS - w), y: Math.max(0, Math.round(r.y)), w, h };
}

/** Sorts items top-to-bottom, then left-to-right. */
export function sortByPosition<T extends Rect>(items: readonly T[]): T[] {
  return [...items].sort((a, b) => a.y - b.y || a.x - b.x);
}

/**
 * Applies gravity: every item moves up as far as possible. The pinned item
 * (e.g. the one being dragged) keeps its position and others flow around it.
 */
export function compact(items: readonly LayoutItem[], pinnedId?: string): LayoutItem[] {
  const placed: LayoutItem[] = [];
  const pinned = items.find((i) => i.id === pinnedId);
  if (pinned) placed.push({ ...pinned });
  for (const item of sortByPosition(items)) {
    if (item.id === pinnedId) continue;
    const candidate = { ...item, y: 0 };
    while (placed.some((p) => collides(p, candidate))) candidate.y++;
    placed.push(candidate);
  }
  return sortByPosition(placed);
}

/** Moves item `id` to (x, y) and lets the other items make room. */
export function moveItem(items: readonly LayoutItem[], id: string, x: number, y: number): LayoutItem[] {
  const moved = items.map((i) => (i.id === id ? { ...i, ...clampRect({ ...i, x, y }) } : i));
  return compact(compact(moved, id));
}

/** Resizes item `id` to w × h (clamped to the grid) and reflows the layout. */
export function resizeItem(items: readonly LayoutItem[], id: string, w: number, h: number): LayoutItem[] {
  const resized = items.map((i) => (i.id === id ? { ...i, ...clampRect({ ...i, w, h }) } : i));
  return compact(compact(resized, id));
}

/** Finds the top-most, left-most position where a w × h item fits. */
export function findFreeSpot(items: readonly Rect[], w: number, h: number): Rect {
  const size = clampRect({ x: 0, y: 0, w, h });
  for (let y = 0; ; y++) {
    for (let x = 0; x + size.w <= GRID_COLUMNS; x++) {
      const candidate = { x, y, w: size.w, h: size.h };
      if (!items.some((i) => collides(i, candidate))) return candidate;
    }
  }
}

/** Reports whether two layouts place every item identically. */
export function sameLayout(a: readonly LayoutItem[], b: readonly LayoutItem[]): boolean {
  if (a.length !== b.length) return false;
  const byId = new Map(b.map((i) => [i.id, i]));
  return a.every((i) => {
    const o = byId.get(i.id);
    return !!o && o.x === i.x && o.y === i.y && o.w === i.w && o.h === i.h;
  });
}
