/** Configuration of a clock widget. */
export interface ClockConfig {
  /** IANA time zone, empty for the browser's zone. */
  timeZone: string;
  /** Optional label, e.g. a city name. */
  label: string;
  hour12: boolean;
  showSeconds: boolean;
  showDate: boolean;
}

/** Defaults for a newly added clock. */
export const defaultClockConfig: ClockConfig = {
  timeZone: '',
  label: '',
  hour12: false,
  showSeconds: false,
  showDate: true,
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
