/** Configuration of a clock widget. */
export interface ClockConfig {
  /** IANA time zone, empty for the browser's zone. */
  timeZone: string;
  /** Optional label, e.g. a city name. */
  label: string;
  hour12: boolean;
  showSeconds: boolean;
  showDate: boolean;
  /** Shows the ISO 8601 calendar week below the date (#122). */
  showWeek: boolean;
}

/** Defaults for a newly added clock. */
export const defaultClockConfig: ClockConfig = {
  timeZone: '',
  label: '',
  hour12: false,
  showSeconds: false,
  showDate: true,
  showWeek: false,
};

/** Reads a stored configuration, falling back to defaults for missing or invalid fields. */
export function parseClockConfig(raw: Record<string, unknown>): ClockConfig {
  const str = (v: unknown, d: string) => (typeof v === 'string' ? v : d);
  const bool = (v: unknown, d: boolean) => (typeof v === 'boolean' ? v : d);
  const timeZone = str(raw['timeZone'], '');
  return {
    timeZone: isValidTimeZone(timeZone) ? timeZone : '',
    label: str(raw['label'], ''),
    hour12: bool(raw['hour12'], defaultClockConfig.hour12),
    showSeconds: bool(raw['showSeconds'], defaultClockConfig.showSeconds),
    showDate: bool(raw['showDate'], defaultClockConfig.showDate),
    showWeek: bool(raw['showWeek'], defaultClockConfig.showWeek),
  };
}

/** Reports whether the runtime knows the IANA time zone (empty means local). */
export function isValidTimeZone(tz: string): boolean {
  if (tz === '') return true;
  try {
    new Intl.DateTimeFormat('en', { timeZone: tz });
    return true;
  } catch {
    return false;
  }
}

/** Formats the time part according to the configuration. */
export function formatTime(date: Date, cfg: ClockConfig, locale?: string): string {
  return new Intl.DateTimeFormat(locale, {
    hour: '2-digit',
    minute: '2-digit',
    second: cfg.showSeconds ? '2-digit' : undefined,
    hour12: cfg.hour12,
    timeZone: cfg.timeZone || undefined,
  }).format(date);
}

/** Formats the date part according to the configuration. */
export function formatDate(date: Date, cfg: ClockConfig, locale?: string): string {
  return new Intl.DateTimeFormat(locale, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    timeZone: cfg.timeZone || undefined,
  }).format(date);
}

const DAY_MS = 86_400_000;

/**
 * ISO 8601 calendar week (1-53) of the day the instant falls on in the
 * configured time zone: weeks start on Monday and week 1 is the one with the
 * year's first Thursday, so early January can belong to the previous year's
 * week 52/53 and late December to week 1 (#122).
 */
export function isoWeek(date: Date, cfg: Pick<ClockConfig, 'timeZone'>): number {
  const parts = new Intl.DateTimeFormat('en-CA', {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    timeZone: cfg.timeZone || undefined,
  }).formatToParts(date);
  const part = (type: string) => Number(parts.find((p) => p.type === type)?.value);
  const day = Date.UTC(part('year'), part('month') - 1, part('day'));
  // Move to the Thursday of this ISO week; its year decides the week-year.
  const weekdayFromMonday = (new Date(day).getUTCDay() + 6) % 7;
  const thursday = day + (3 - weekdayFromMonday) * DAY_MS;
  const firstOfYear = Date.UTC(new Date(thursday).getUTCFullYear(), 0, 1);
  return Math.floor((thursday - firstOfYear) / DAY_MS / 7) + 1;
}
