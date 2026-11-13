import type { Action, Entity } from './intent';

/**
 * Answers of the smart search as data (#156). Five kinds are rendered by
 * generic components; a new kind is added only when at least three
 * sources need it. Severity and status always come from the backend –
 * the frontend only shows them.
 */

export type Severity = 'ok' | 'warning' | 'critical';
export type State = 'ok' | 'warning' | 'error' | 'off';

interface Base {
  /** Intent that produced the answer. */
  intent: string;
  title: string;
  /** Widget kind with the details ("open widget"). */
  widgetKind?: string;
  /** When the data was measured or fetched (#94). */
  at: Date | null;
  /** Plain value for the clipboard; Enter copies it (calculator, #160). */
  copy?: string;
}

export interface MetricAnswer extends Base {
  kind: 'metric';
  value: string;
  detail?: string;
  /** 0–100 for a meter, if the value is a share. */
  percent?: number;
  severity?: Severity;
}

export interface StatusAnswer extends Base {
  kind: 'status';
  state: State;
  text: string;
  detail?: string;
}

export interface ListAnswer extends Base {
  kind: 'list';
  items: { label: string; value: string; state?: State; percent?: number }[];
  /** Shown when the list is empty, e.g. "all services online". */
  empty: string;
}

export interface TextAnswer extends Base {
  kind: 'text';
  text: string;
  /** Names to pick from when a slot is missing (#158). */
  suggestions?: string[];
}

/** A write action; always confirmed and never offered to guests. */
export interface ActionItem {
  source: string;
  action: Action;
  target: Entity;
  label: string;
  confirm: string;
}

export interface ActionsAnswer extends Base {
  kind: 'actions';
  text: string;
  actions: ActionItem[];
}

export type Answer = MetricAnswer | StatusAnswer | ListAnswer | TextAnswer | ActionsAnswer;

/** Age of the data of an answer (#94) as a message key and number. */
export function dataAge(
  at: Date | null,
  now: Date,
): { key: 'smart.age_now' | 'smart.age_min' | 'smart.age_hours'; n: number } | null {
  if (!at) return null;
  const minutes = Math.max(0, Math.floor((now.getTime() - at.getTime()) / 60_000));
  if (minutes < 1) return { key: 'smart.age_now', n: 0 };
  if (minutes < 120) return { key: 'smart.age_min', n: minutes };
  return { key: 'smart.age_hours', n: Math.floor(minutes / 60) };
}
