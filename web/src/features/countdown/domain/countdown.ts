/** A countdown to a day, optionally with a time of day (#54). */
export interface Countdown {
  id: string;
  title: string;
  emoji: string;
  /** YYYY-MM-DD */
  date: string;
  /** HH:MM, or '' for the whole day. */
  time: string;
}

/** What the user enters. */
export type CountdownInput = Omit<Countdown, 'id'>;

/** Limits; mirror the backend. */
export const MAX_TITLE_LENGTH = 120;
export const MAX_GRACE_DAYS = 365;

/** Per-widget settings. */
export interface CountdownConfig {
  /** Countdowns with fewer days left are highlighted. */
  highlightDays: number;
  /** Past countdowns stay visible for this many days. */
  graceDays: number;
}

export const defaultCountdownConfig: CountdownConfig = { highlightDays: 7, graceDays: 1 };

const int = (v: unknown, d: number, lo: number, hi: number) =>
  typeof v === 'number' && Number.isInteger(v) ? Math.min(Math.max(v, lo), hi) : d;

/** Reads the widget configuration, falling back to defaults. */
export function parseCountdownConfig(raw: Record<string, unknown>): CountdownConfig {
  return {
    highlightDays: int(raw['highlightDays'], defaultCountdownConfig.highlightDays, 0, 365),
    graceDays: int(raw['graceDays'], defaultCountdownConfig.graceDays, 0, MAX_GRACE_DAYS),
  };
}

/** The moment a countdown ends, in the browser's time zone. */
export function target(c: Pick<Countdown, 'date' | 'time'>): Date {
  const [y = 0, m = 1, d = 1] = c.date.split('-').map(Number);
  const [hh = 0, mm = 0] = c.time ? c.time.split(':').map(Number) : [];
  return new Date(y, m - 1, d, hh, mm);
}

/** Remaining time of a countdown at a moment. */
export type Remaining =
  /** Ended: a whole-day countdown on its day, or a timed one after its time. */
  | { kind: 'today' }
  | { kind: 'past' }
  /** Less than a day left of a timed countdown. */
  | { kind: 'hours'; hours: number; minutes: number }
  /** Calendar days until the day. */
  | { kind: 'days'; days: number };

const DAY = 24 * 60 * 60 * 1000;

/** Calendar days between two dates (ignores the time of day, DST-safe). */
function dayDiff(from: Date, to: Date): number {
  const a = Date.UTC(from.getFullYear(), from.getMonth(), from.getDate());
  const b = Date.UTC(to.getFullYear(), to.getMonth(), to.getDate());
  return Math.round((b - a) / DAY);
}

/**
 * Remaining time at now: days for countdowns on later days, hours and
 * minutes on the last day of a timed countdown.
 */
export function remainingAt(c: Pick<Countdown, 'date' | 'time'>, now: Date): Remaining {
  const end = target(c);
  const days = dayDiff(now, end);
  if (!c.time) {
    if (days > 0) return { kind: 'days', days };
    return days === 0 ? { kind: 'today' } : { kind: 'past' };
  }
  const ms = end.getTime() - now.getTime();
  if (ms <= 0) return { kind: 'past' };
  if (ms < DAY) {
    const total = Math.ceil(ms / 60_000);
    return { kind: 'hours', hours: Math.floor(total / 60), minutes: total % 60 };
  }
  return { kind: 'days', days };
}

/** Whether a countdown is close enough to be highlighted. */
export function isSoon(r: Remaining, highlightDays: number): boolean {
  return r.kind === 'today' || r.kind === 'hours' || (r.kind === 'days' && r.days < highlightDays);
}

/** Validates the input; returns a message key or null. */
export function validateInput(
  i: CountdownInput,
): 'countdown.invalid_title' | 'countdown.invalid_date' | 'countdown.invalid_time' | null {
  const title = i.title.trim();
  if (title === '' || [...title].length > MAX_TITLE_LENGTH) return 'countdown.invalid_title';
  if (!/^\d{4}-\d{2}-\d{2}$/.test(i.date) || Number.isNaN(target({ date: i.date, time: '' }).getTime()))
    return 'countdown.invalid_date';
  if (i.time && !/^([01]\d|2[0-3]):[0-5]\d$/.test(i.time)) return 'countdown.invalid_time';
  return null;
}
