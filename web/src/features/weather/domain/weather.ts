/** Configuration of a weather widget. */
export interface WeatherConfig {
  lat: number;
  lon: number;
  label: string;
  /** Number of forecast days to show (0–5). */
  days: number;
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
  return { lat, lon, label: typeof raw['label'] === 'string' ? raw['label'] : '', days };
}

/** Human readable label and emoji for a WMO weather code. */
export interface Condition {
  label: string;
  icon: string;
}

const conditions: Record<number, Condition> = {
  0: { label: 'Clear sky', icon: '☀️' },
  1: { label: 'Mainly clear', icon: '🌤️' },
  2: { label: 'Partly cloudy', icon: '⛅' },
  3: { label: 'Overcast', icon: '☁️' },
  45: { label: 'Fog', icon: '🌫️' },
  48: { label: 'Rime fog', icon: '🌫️' },
  51: { label: 'Light drizzle', icon: '🌦️' },
  53: { label: 'Drizzle', icon: '🌦️' },
  55: { label: 'Dense drizzle', icon: '🌧️' },
  56: { label: 'Freezing drizzle', icon: '🌧️' },
  57: { label: 'Freezing drizzle', icon: '🌧️' },
  61: { label: 'Light rain', icon: '🌦️' },
  63: { label: 'Rain', icon: '🌧️' },
  65: { label: 'Heavy rain', icon: '🌧️' },
  66: { label: 'Freezing rain', icon: '🌧️' },
  67: { label: 'Freezing rain', icon: '🌧️' },
  71: { label: 'Light snow', icon: '🌨️' },
  73: { label: 'Snow', icon: '🌨️' },
  75: { label: 'Heavy snow', icon: '❄️' },
  77: { label: 'Snow grains', icon: '🌨️' },
  80: { label: 'Rain showers', icon: '🌦️' },
  81: { label: 'Rain showers', icon: '🌧️' },
  82: { label: 'Violent showers', icon: '⛈️' },
  85: { label: 'Snow showers', icon: '🌨️' },
  86: { label: 'Snow showers', icon: '❄️' },
  95: { label: 'Thunderstorm', icon: '⛈️' },
  96: { label: 'Thunderstorm with hail', icon: '⛈️' },
  99: { label: 'Thunderstorm with hail', icon: '⛈️' },
};

/** Maps a WMO weather code to a label and icon; night variants for clear skies. */
export function describe(code: number, isDay = true): Condition {
  const c = conditions[code] ?? { label: 'Unknown', icon: '❔' };
  if (!isDay && (code === 0 || code === 1)) return { ...c, icon: '🌙' };
  return c;
}

/** Formats a place for display, e.g. "Oldenburg, Lower Saxony, Germany". */
export function placeLabel(p: Place): string {
  return [p.name, p.region, p.country]
    .filter((s) => s && s !== p.name)
    .reduce((a, s) => `${a}, ${s}`, p.name);
}
