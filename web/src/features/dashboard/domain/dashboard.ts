import type { WidgetCategory } from './catalog';
import { t } from '../../../shared/i18n';
import type { LayoutItem } from './layout';
import type { MobileItem } from './mobile';

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
  /** Rules that open this tab by default (#71). */
  rules: TabRule[];
  widgets: Widget[];
  /** Phone layout in display order (#92). */
  mobile: MobileItem[];
  /** False while the phone layout follows the desktop grid. */
  mobileCustom: boolean;
}

/**
 * Opens a tab by default (#71). Empty parts always match: no days = every
 * day, empty from/to = whole day, no networks = any network. Evaluated by
 * the server, which knows the client address.
 */
export interface TabRule {
  /** Weekdays, 0 = Sunday. */
  days: number[];
  /** "HH:MM"; a range like 22:00–02:00 runs over midnight. */
  from: string;
  to: string;
  /** CIDR ranges such as 192.168.178.0/24. */
  networks: string[];
}

/** Limits of tab rules; mirror the backend. */
export const MAX_TAB_RULES = 10;
export const MAX_RULE_NETWORKS = 10;

const TIME = /^([01]\d|2[0-3]):[0-5]\d$/;
const CIDR4 = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})\/(\d{1,2})$/;
const CIDR6 = /^[0-9a-f:.]+\/(\d{1,3})$/i;

/** Rough CIDR check; the backend validates exactly. */
export function isCidr(value: string): boolean {
  const v4 = CIDR4.exec(value);
  if (v4) return v4.slice(1, 5).every((p) => Number(p) <= 255) && Number(v4[5]) <= 32;
  const v6 = CIDR6.exec(value);
  return v6 !== null && value.includes(':') && Number(v6[1]) <= 128;
}

/** Splits a comma- or whitespace-separated list of networks. */
export function parseNetworks(text: string): string[] {
  return text
    .split(/[\s,]+/)
    .map((s) => s.trim())
    .filter((s) => s !== '');
}

/** Validates rules and returns an error message or null. */
export function validateRules(rules: readonly TabRule[]): string | null {
  if (rules.length > MAX_TAB_RULES) return t('dashboard.rules_too_many', { max: MAX_TAB_RULES });
  for (const r of rules) {
    if ((r.from === '') !== (r.to === '')) return t('dashboard.rule_needs_both_times');
    if ((r.from !== '' && !TIME.test(r.from)) || (r.to !== '' && !TIME.test(r.to)))
      return t('dashboard.rule_invalid_time');
    if (r.networks.length > MAX_RULE_NETWORKS)
      return t('dashboard.rule_too_many_networks', { max: MAX_RULE_NETWORKS });
    const bad = r.networks.find((n) => !isCidr(n));
    if (bad !== undefined) return t('dashboard.rule_invalid_network', { network: bad });
  }
  return null;
}

/**
 * Picks the tab to show on load (#71): a tab chosen by hand in this
 * browser session wins, then the tab whose rules match, then the last
 * used tab, then the first one. IDs of deleted tabs are ignored.
 */
export function initialTabId(
  tabIds: readonly string[],
  choice: { manual: string; byRule: string; lastUsed: string },
): string {
  for (const id of [choice.manual, choice.byRule, choice.lastUsed]) {
    if (id !== '' && tabIds.includes(id)) return id;
  }
  return tabIds[0] ?? '';
}

/** All tabs plus the tab whose rules match right now ("" = none, #71). */
export interface Dashboard {
  tabs: Tab[];
  defaultTabId: string;
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
