import type { WidgetCategory } from './catalog';
import { t } from '../../../shared/i18n';
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
  /** Background image URL of this tab ("" = global background). */
  background: string;
  widgets: Widget[];
}

/** A saved tab layout new tabs can be created from. */
export interface TabTemplate {
  id: string;
  name: string;
  widgets: number;
  kinds: string[];
}

/** Validates a background URL (empty is allowed) and returns an error or null. */
export function validateBackground(url: string): string | null {
  const v = url.trim();
  return v === '' || /^https?:\/\/\S+$/i.test(v) ? null : t('dashboard.invalid_background');
}

/** Describes a widget kind that can be added to a tab. */
export interface WidgetType {
  kind: string;
  title: string;
  description: string;
  defaultSize: { w: number; h: number };
  /** Column of the widget picker (#104). */
  category: WidgetCategory;
  /** Extra search terms of the widget picker. */
  keywords: string[];
}

/** Maximum tab name length; mirrors the backend. */
export const MAX_TAB_NAME_LENGTH = 40;

/** Validates a tab name and returns an error message or null. */
export function validateTabName(name: string): string | null {
  const trimmed = name.trim();
  if (trimmed === '') return t('dashboard.tab_name_empty');
  if ([...trimmed].length > MAX_TAB_NAME_LENGTH)
    return t('dashboard.tab_name_too_long', { max: MAX_TAB_NAME_LENGTH });
  return null;
}

/** Moves the element at `from` to index `to` and returns a new array. */
export function reorder<T>(list: readonly T[], from: number, to: number): T[] {
  const out = [...list];
  const [moved] = out.splice(from, 1);
  if (moved !== undefined) out.splice(Math.max(0, Math.min(to, out.length)), 0, moved);
  return out;
}
