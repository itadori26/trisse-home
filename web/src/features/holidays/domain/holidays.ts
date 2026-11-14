import { t, type MessageKey } from '../../../shared/i18n';

/** German federal states (ISO 3166-2:DE codes). */
export const STATES = [
  'BW',
  'BY',
  'BE',
  'BB',
  'HB',
  'HH',
  'HE',
  'MV',
  'NI',
  'NW',
  'RP',
  'SL',
  'SN',
  'ST',
  'SH',
  'TH',
] as const;
export type State = (typeof STATES)[number];
export const DEFAULT_STATE: State = 'NI';

/** A public holiday or a school holiday period (#55). */
export interface HolidayEntry {
  kind: 'public' | 'school';
  /** Public holiday key, e.g. good_friday. */
  key: string;
  /** Name of a school holiday period. */
  name: string;
  start: Date;
  /** Inclusive. */
  end: Date;
  current: boolean;
  daysUntil: number;
}

export interface HolidaysOverview {
  state: State;
  entries: HolidayEntry[];
  school: boolean;
  schoolUpdatedAt: Date | null;
  schoolError: string;
}

/** Configuration of the widget. */
export interface HolidaysConfig {
  state: State;
  count: number;
}

/** Reads a stored widget configuration. */
export function parseHolidaysConfig(raw: Record<string, unknown>): HolidaysConfig {
  const state = STATES.includes(raw['state'] as State) ? (raw['state'] as State) : DEFAULT_STATE;
  const count = typeof raw['count'] === 'number' ? Math.round(raw['count']) : 6;
  return { state, count: Math.min(Math.max(count, 1), 20) };
}

/** The display name of an entry: translated public holiday or the school holiday name. */
export function entryName(e: Pick<HolidayEntry, 'kind' | 'key' | 'name'>): string {
  if (e.kind === 'school') return e.name;
  const key = `holidays.day.${e.key}` as MessageKey;
  const name = t(key);
  return name === key ? e.key : name;
}

/** Name of a state. */
export const stateName = (s: State): string => t(`holidays.state.${s}` as MessageKey);

/** Whole days between two dates, for a period's length (inclusive). */
export function periodDays(e: Pick<HolidayEntry, 'start' | 'end'>): number {
  return Math.round((e.end.getTime() - e.start.getTime()) / 86_400_000) + 1;
}
