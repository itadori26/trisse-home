import type { LayoutItem } from './layout';

/** Opaque widget configuration owned by the widget's feature. */
export type WidgetConfig = Record<string, unknown>;

/** A widget placed on a tab. */
export interface Widget extends LayoutItem {
  tabId: string;
  kind: string;
  config: WidgetConfig;
}

/** A dashboard tab with its widgets. */
export interface Tab {
  id: string;
  name: string;
  position: number;
  widgets: Widget[];
}

/** Describes a widget kind that can be added to a tab. */
export interface WidgetType {
  kind: string;
  title: string;
  description: string;
  defaultSize: { w: number; h: number };
}

/** Maximum tab name length; mirrors the backend. */
export const MAX_TAB_NAME_LENGTH = 40;

/** Validates a tab name and returns an error message or null. */
export function validateTabName(name: string): string | null {
  const trimmed = name.trim();
  if (trimmed === '') return 'The tab name must not be empty.';
  if ([...trimmed].length > MAX_TAB_NAME_LENGTH) return 'The tab name must be at most 40 characters long.';
  return null;
}

/** Moves the element at `from` to index `to` and returns a new array. */
export function reorder<T>(list: readonly T[], from: number, to: number): T[] {
  const out = [...list];
  const [moved] = out.splice(from, 1);
  if (moved !== undefined) out.splice(Math.max(0, Math.min(to, out.length)), 0, moved);
  return out;
}
