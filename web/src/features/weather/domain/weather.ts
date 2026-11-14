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
  /** DWD pollen region (0 = no pollen section, #56). */
  pollenRegion: number;
  /** Pollen types to show (empty = all). */
  pollenTypes: PollenType[];
  /** DWD UV index city ('' = no UV index). */
  uvCity: string;
}

/** The eight allergens of the DWD pollen forecast, in season order. */
export const POLLEN_TYPES = [
  'hazel',
  'alder',
  'ash',
  'birch',
  'grasses',
  'rye',
  'mugwort',
  'ambrosia',
] as const;
export type PollenType = (typeof POLLEN_TYPES)[number];

export type PollenBurden = 'none' | 'low' | 'moderate' | 'high' | 'unknown';
export type UVLevel = 'low' | 'moderate' | 'high' | 'very_high' | 'extreme' | 'unknown';

/** Pollen forecast of one allergen; levels in DWD notation ("0-1"), '' = no data. */
export interface PollenDay {
  type: PollenType;
  today: string;
  tomorrow: string;
  todayBurden: PollenBurden;
  tomorrowBurden: PollenBurden;
}

/** Pollen and UV index of a widget; a part is null when not configured or unknown. */
export interface HealthReport {
  pollen: { regionName: string; days: PollenDay[] } | null;
  uv: { city: string; today: number; tomorrow: number; todayLevel: UVLevel; tomorrowLevel: UVLevel } | null;
}

/** Selectable DWD pollen regions and UV cities. */
export interface HealthOptions {
  regions: { id: number; name: string }[];
  cities: string[];
}

/** The config fields of a widget without pollen and UV index. */
export const noHealth: Pick<WeatherConfig, 'pollenRegion' | 'pollenTypes' | 'uvCity'> = {
  pollenRegion: 0,
  pollenTypes: [],
  uvCity: '',
};

/** Whether the widget shows the pollen/UV section at all. */
export const wantsHealth = (c: Pick<WeatherConfig, 'pollenRegion' | 'uvCity'>): boolean =>
  c.pollenRegion > 0 || c.uvCity !== '';

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
  /** Degrees the wind comes from (0 = north, 90 = east); null when unknown (#118). */
  windDirection: number | null;
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
    pollenRegion: typeof raw['pollenRegion'] === 'number' ? Math.max(0, Math.round(raw['pollenRegion'])) : 0,
    pollenTypes: Array.isArray(raw['pollenTypes'])
      ? POLLEN_TYPES.filter((p) => (raw['pollenTypes'] as unknown[]).includes(p))
      : [],
    uvCity: typeof raw['uvCity'] === 'string' ? raw['uvCity'] : '',
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

/** The eight points of the compass rose. */
export type Compass = 'n' | 'ne' | 'e' | 'se' | 's' | 'sw' | 'w' | 'nw';

const COMPASS: readonly Compass[] = ['n', 'ne', 'e', 'se', 's', 'sw', 'w', 'nw'];

/**
 * The compass point a direction in degrees belongs to. Each point covers 45°
 * centred on itself, so north is [337.5°, 22.5°). Angles outside 0-360 wrap
 * around (-45 is north-west).
 */
export function compass(degrees: number): Compass {
  const normalised = ((degrees % 360) + 360) % 360;
  return COMPASS[Math.round(normalised / 45) % 8] ?? 'n';
}
