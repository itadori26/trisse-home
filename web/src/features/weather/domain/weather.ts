import { intlLocale, type MessageKey, t } from '../../../shared/i18n';
/** Configuration of a weather widget. */
export interface WeatherConfig {
  lat: number;
  lon: number;
  label: string;
  /** Number of forecast days to show (0–5). */
  days: number;
  /** Show official warnings (DWD, Germany only). */
  warnings: boolean;
}

/** One forecast day. */
export interface ForecastDay {
  date: Date;
  min: number;
  max: number;
  code: number;
  precipitationProbability: number;
}

/** Current conditions and outlook. */
export interface Forecast {
  temperature: number;
  apparentTemperature: number;
  humidity: number;
  windSpeed: number;
  code: number;
  isDay: boolean;
  days: ForecastDay[];
  updatedAt: Date;
}

/** A geocoding result. */
export interface Place {
  name: string;
  region: string;
  country: string;
  lat: number;
  lon: number;
}

/** Reads a stored widget configuration; returns null if no location is set. */
export function parseWeatherConfig(raw: Record<string, unknown>): WeatherConfig | null {
  const lat = raw['lat'];
  const lon = raw['lon'];
  if (typeof lat !== 'number' || typeof lon !== 'number') return null;
  const days = typeof raw['days'] === 'number' ? Math.max(0, Math.min(5, Math.round(raw['days']))) : 3;
  return {
    lat,
    lon,
    label: typeof raw['label'] === 'string' ? raw['label'] : '',
    days,
    warnings: raw['warnings'] !== false,
  };
}

/** Human readable label and emoji for a WMO weather code. */
export interface Condition {
  label: string;
  icon: string;
}

const conditions: Record<number, Condition> = {
  0: { label: 'weather.cond.clear_sky', icon: '☀️' },
  1: { label: 'weather.cond.mainly_clear', icon: '🌤️' },
  2: { label: 'weather.cond.partly_cloudy', icon: '⛅' },
  3: { label: 'weather.cond.overcast', icon: '☁️' },
  45: { label: 'weather.cond.fog', icon: '🌫️' },
  48: { label: 'weather.cond.rime_fog', icon: '🌫️' },
  51: { label: 'weather.cond.light_drizzle', icon: '🌦️' },
  53: { label: 'weather.cond.drizzle', icon: '🌦️' },
  55: { label: 'weather.cond.dense_drizzle', icon: '🌧️' },
  56: { label: 'weather.cond.freezing_drizzle', icon: '🌧️' },
  57: { label: 'weather.cond.freezing_drizzle', icon: '🌧️' },
  61: { label: 'weather.cond.light_rain', icon: '🌦️' },
  63: { label: 'weather.cond.rain', icon: '🌧️' },
  65: { label: 'weather.cond.heavy_rain', icon: '🌧️' },
  66: { label: 'weather.cond.freezing_rain', icon: '🌧️' },
  67: { label: 'weather.cond.freezing_rain', icon: '🌧️' },
  71: { label: 'weather.cond.light_snow', icon: '🌨️' },
  73: { label: 'weather.cond.snow', icon: '🌨️' },
  75: { label: 'weather.cond.heavy_snow', icon: '❄️' },
  77: { label: 'weather.cond.snow_grains', icon: '🌨️' },
  80: { label: 'weather.cond.rain_showers', icon: '🌦️' },
  81: { label: 'weather.cond.rain_showers', icon: '🌧️' },
  82: { label: 'weather.cond.violent_showers', icon: '⛈️' },
  85: { label: 'weather.cond.snow_showers', icon: '🌨️' },
  86: { label: 'weather.cond.snow_showers', icon: '❄️' },
  95: { label: 'weather.cond.thunderstorm', icon: '⛈️' },
  96: { label: 'weather.cond.thunderstorm_with_hail', icon: '⛈️' },
  99: { label: 'weather.cond.thunderstorm_with_hail', icon: '⛈️' },
};

/** Maps a WMO weather code to a label and icon; night variants for clear skies. */
export function describe(code: number, isDay = true): Condition {
  const c = conditions[code] ?? { label: 'weather.cond.unknown', icon: '❔' };
  const label = t(c.label as MessageKey);
  if (!isDay && (code === 0 || code === 1)) return { label, icon: '🌙' };
  return { label, icon: c.icon };
}

/** Formats a place for display, e.g. "Oldenburg, Lower Saxony, Germany". */
export function placeLabel(p: Place): string {
  return [p.name, p.region, p.country]
    .filter((s) => s && s !== p.name)
    .reduce((a, s) => `${a}, ${s}`, p.name);
}

/** Severity of an official warning. */
export type WarningSeverity = 'minor' | 'moderate' | 'severe' | 'extreme';

/** An official weather warning. */
export interface WeatherWarning {
  id: string;
  event: string;
  headline: string;
  description: string;
  instruction: string;
  severity: WarningSeverity;
  onset: Date | null;
  expires: Date | null;
}

/** DWD warning colors per severity (levels 1–4). */
export const severityColors: Record<WarningSeverity, string> = {
  minor: '#ffeb3b',
  moderate: '#fb8c00',
  severe: '#e53935',
  extreme: '#880e4f',
};

/** Validity text such as "until Thu 17:00" or "from Fri 09:00". */
export function validity(w: Pick<WeatherWarning, 'onset' | 'expires'>, now: Date): string {
  const fmt = (d: Date) =>
    d.toLocaleString(intlLocale(), { weekday: 'short', hour: '2-digit', minute: '2-digit' });
  if (w.onset && w.onset > now) return t('weather.valid_from', { time: fmt(w.onset) });
  if (w.expires) return t('weather.valid_until', { time: fmt(w.expires) });
  return '';
}

/** Cache-busting key that changes every five minutes (radar refresh). */
export function radarKey(now: Date): number {
  return Math.floor(now.getTime() / 300_000);
}
