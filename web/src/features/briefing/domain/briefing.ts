import { t, type MessageKey } from '../../../shared/i18n';

/** Parts of the briefing, in their default order (#52). */
export const SECTION_KINDS = ['warnings', 'incidents', 'services', 'todos', 'events', 'waste'] as const;
export type SectionKind = (typeof SECTION_KINDS)[number];

export type Severity = 'info' | 'warning' | 'critical';

/** One line of a section. */
export interface BriefingItem {
  text: string;
  detail: string;
  /** Time of the item (due date, start, collection day …) or null. */
  at: Date | null;
  severity: Severity;
}

export interface BriefingSection {
  kind: SectionKind;
  items: BriefingItem[];
  /** Items that did not fit. */
  more: number;
}

/** The composed briefing of today. */
export interface Briefing {
  sections: BriefingSection[];
  /** Sections whose source could not be read right now. */
  unavailable: SectionKind[];
}

export interface LayoutEntry {
  kind: SectionKind;
  enabled: boolean;
}

/** Order and visibility of the sections, daily push and weather location. */
export interface BriefingSettings {
  layout: LayoutEntry[];
  pushOn: boolean;
  /** "HH:MM" */
  pushAt: string;
  /** Weekdays, 0 = Sunday; empty = every day. */
  pushDays: number[];
  skipEmpty: boolean;
  lat: number;
  lon: number;
  /** Sections whose source module is enabled. */
  available: SectionKind[];
}

export const isSectionKind = (k: string): k is SectionKind =>
  (SECTION_KINDS as readonly string[]).includes(k);

const titles: Record<SectionKind, MessageKey> = {
  warnings: 'briefing.section.warnings',
  incidents: 'briefing.section.incidents',
  services: 'briefing.section.services',
  todos: 'briefing.section.todos',
  events: 'briefing.section.events',
  waste: 'briefing.section.waste',
};

/** Heading of a section. */
export const sectionTitle = (kind: SectionKind): string => t(titles[kind]);

/** Widget kinds that show the source of a section, in order of preference. */
export const SOURCE_WIDGETS: Record<SectionKind, string[]> = {
  warnings: ['weather', 'radar'],
  incidents: ['timeline'],
  services: ['services', 'status-overview'],
  todos: ['todos'],
  events: ['calendar'],
  waste: ['waste'],
};

/** Moves a layout entry by delta places; returns null at the edges. */
export function moveEntry(
  layout: readonly LayoutEntry[],
  index: number,
  delta: number,
): LayoutEntry[] | null {
  const to = index + delta;
  if (index < 0 || index >= layout.length || to < 0 || to >= layout.length) return null;
  const out = [...layout];
  const [moved] = out.splice(index, 1);
  if (moved) out.splice(to, 0, moved);
  return out;
}

/** Validates the settings before saving; returns an error message or null. */
export function validateSettings(s: BriefingSettings): string | null {
  if (!/^([01]\d|2[0-3]):[0-5]\d$/.test(s.pushAt)) return t('briefing.invalid_time');
  if (
    !Number.isFinite(s.lat) ||
    s.lat < -90 ||
    s.lat > 90 ||
    !Number.isFinite(s.lon) ||
    s.lon < -180 ||
    s.lon > 180
  )
    return t('briefing.invalid_location');
  return null;
}

/** Markers the server sends as detail of to-dos and collections. */
const detailKeys: Record<string, MessageKey> = {
  overdue: 'briefing.detail.overdue',
  today: 'briefing.detail.today',
  tomorrow: 'briefing.detail.tomorrow',
};

/**
 * Secondary text of an item: the markers of to-dos and collections are
 * translated, free text (locations, error messages) is shown as is.
 */
export function itemDetail(kind: SectionKind, item: BriefingItem): string {
  const key = kind === 'todos' || kind === 'waste' ? detailKeys[item.detail] : undefined;
  return key ? t(key) : item.detail;
}

/**
 * Short time label of an item: the clock time for appointments and
 * incidents, nothing for date-only items (their detail says enough).
 */
export function itemTime(kind: SectionKind, item: BriefingItem, locale?: string): string {
  if (!item.at || (kind !== 'events' && kind !== 'incidents')) return '';
  return item.at.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' });
}
