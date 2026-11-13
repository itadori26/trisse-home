import { t } from '../../../shared/i18n';

/** Fuel types (#53). */
export type FuelType = 'e5' | 'e10' | 'diesel';

/** Every fuel type, in the order of the settings. */
export const fuelTypes: readonly FuelType[] = ['e5', 'e10', 'diesel'];

/** Display name of a fuel type. */
export function fuelName(f: FuelType): string {
  return { e5: 'Super E5', e10: 'Super E10', diesel: 'Diesel' }[f];
}

/** A favourite station. */
export interface Favourite {
  id: string;
  name: string;
}

/** Settings as shown in the UI (the API key is never returned). */
export interface FuelSettings {
  configured: boolean;
  hasApiKey: boolean;
  lat: number;
  lon: number;
  place: string;
  radiusKm: number;
  fuel: FuelType;
  favourites: Favourite[];
}

/** A station with the price of the chosen fuel. */
export interface Station {
  id: string;
  name: string;
  brand: string;
  address: string;
  distanceKm: number;
  /** Null when the station reports no price. */
  price: number | null;
  open: boolean;
}

/** What the widget shows. */
export interface FuelOverview {
  fuel: FuelType;
  place: string;
  cheapest: Station[];
  favourites: Station[];
}

/** A place found by the (weather) place search. */
export interface PlaceResult {
  name: string;
  region: string;
  lat: number;
  lon: number;
}

/** German pump notation: 1.739 → "1,73⁹"; the last digit is superscript. */
export function formatPrice(price: number | null, locale: string): { main: string; tenth: string } | null {
  if (price === null) return null;
  const fixed = price.toFixed(3);
  const main = new Intl.NumberFormat(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(
    Math.floor(price * 100) / 100,
  );
  return { main, tenth: fixed.slice(-1) };
}

/** Whether a station is one of the favourites. */
export function isFavourite(favs: readonly Favourite[], id: string): boolean {
  return favs.some((f) => f.id === id);
}

/** Adds or removes a station from the favourites (at most ten). */
export function toggleFavourite(
  favs: readonly Favourite[],
  s: Pick<Station, 'id' | 'name' | 'brand'>,
): Favourite[] {
  if (isFavourite(favs, s.id)) return favs.filter((f) => f.id !== s.id);
  return [...favs, { id: s.id, name: s.brand || s.name }].slice(0, 10);
}

/** Widget configuration: which list to show. */
export interface FuelWidgetConfig {
  view: 'cheapest' | 'favourites';
}

/** Reads the widget configuration. */
export function parseWidgetConfig(raw: Record<string, unknown>): FuelWidgetConfig {
  return { view: raw['view'] === 'favourites' ? 'favourites' : 'cheapest' };
}

/** Validates the settings form. */
export function validateSettings(lat: number, lon: number, radiusKm: number): string | null {
  if (!Number.isFinite(lat) || !Number.isFinite(lon) || (lat === 0 && lon === 0))
    return t('fuel.choose_location');
  if (!(radiusKm >= 1 && radiusKm <= 25)) return t('fuel.invalid_radius');
  return null;
}
