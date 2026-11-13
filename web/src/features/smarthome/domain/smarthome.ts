import { t } from '../../../shared/i18n';

/** Connection settings as shown in the UI (the token is never returned). */
export interface SmartHomeSettings {
  configured: boolean;
  url: string;
  hasToken: boolean;
}

/** What a tile can do; empty when read-only or not allow-listed. */
export type EntityAction = '' | 'toggle' | 'press';

/** The state of one device or sensor. */
export interface Entity {
  id: string;
  name: string;
  state: string;
  unit: string;
  lastChanged: Date | null;
  action: EntityAction;
}

/** Most entities a widget shows, as enforced by the backend. */
export const MAX_ENTITIES = 50;

/** Widget configuration: the entities in display order. */
export interface SmartHomeWidgetConfig {
  entities: string[];
}

/** Reads the widget configuration. */
export function parseWidgetConfig(raw: Record<string, unknown>): SmartHomeWidgetConfig {
  const list = Array.isArray(raw['entities']) ? raw['entities'] : [];
  return { entities: list.filter((x): x is string => typeof x === 'string').slice(0, MAX_ENTITIES) };
}

/** Whether a toggleable entity is on. */
export function isOn(e: Pick<Entity, 'state'>): boolean {
  return e.state === 'on';
}

/** State text with unit, e.g. "42.5 W"; unavailable entities say so. */
export function stateText(e: Pick<Entity, 'state' | 'unit'>): string {
  if (e.state === 'unavailable' || e.state === 'unknown') return t('smarthome.unavailable');
  if (e.state === 'on') return t('smarthome.on');
  if (e.state === 'off') return t('smarthome.off');
  return e.unit ? `${e.state} ${e.unit}` : e.state;
}

/** "5 min ago", "3 h ago" or "2 d ago". */
export function changedAgo(d: Date | null, now: Date): string {
  if (!d) return '';
  const minutes = Math.max(0, (now.getTime() - d.getTime()) / 60_000);
  if (minutes < 60) return t('time.minutes_ago', { n: Math.round(minutes) });
  if (minutes < 48 * 60) return t('time.hours_ago', { n: Math.round(minutes / 60) });
  return t('time.days_ago', { n: Math.round(minutes / 1440) });
}

/** Entities matching a search in ID or name, for the picker. */
export function filterEntities<T extends Pick<Entity, 'id' | 'name'>>(
  list: readonly T[],
  query: string,
): T[] {
  const q = query.trim().toLowerCase();
  if (!q) return [...list];
  return list.filter((e) => e.id.includes(q) || e.name.toLowerCase().includes(q));
}
