/**
 * Saved places (#591) as widgets see them: a place is entered once and
 * chosen by every widget that needs a location. Widgets copy the
 * coordinates into their settings and keep the place's id next to them, so
 * they go on working when the place is deleted and Settings can say which
 * widgets use which place.
 */

/** A place the account saved. */
export interface SavedPlace {
  id: string;
  name: string;
  /** Address line from the search; empty for a browser location. */
  label: string;
  lat: number;
  lon: number;
}

/** What a widget settings form needs: list and save places. */
export interface SavedPlaces {
  list(): Promise<SavedPlace[]>;
  add(place: Omit<SavedPlace, 'id'>): Promise<SavedPlace>;
}

/** The setting under which a widget stores the place it uses. */
export const PLACE_ID_KEY = 'placeId';

/** The place id stored in a widget's settings, or ''. */
export function placeIdOf(config: Record<string, unknown>): string {
  const id = config[PLACE_ID_KEY];
  return typeof id === 'string' ? id : '';
}

/**
 * Finds a saved place at the given coordinates (within about 100 m), so a
 * widget set up before places existed shows its place as chosen.
 */
export function placeAt(places: readonly SavedPlace[], lat: number, lon: number): SavedPlace | undefined {
  const near = 0.001;
  return places.find((p) => Math.abs(p.lat - lat) < near && Math.abs(p.lon - lon) < near);
}

/**
 * Decimal places kept of a browser location: three are about 100 m, enough
 * for weather, stops and fuel prices without storing where someone lives
 * to the metre (#693).
 */
const LOCATION_DECIMALS = 3;

/** Rounds a coordinate to LOCATION_DECIMALS places. */
export function roundCoordinate(value: number): number {
  const f = 10 ** LOCATION_DECIMALS;
  return Math.round(value * f) / f;
}

/**
 * Reads the browser's location, rounded with roundCoordinate; rejects when
 * it is denied or missing. Call it only from a click: the browser asks for
 * permission, and Waylight never reads the location on its own.
 */
export function browserLocation(timeout = 10_000): Promise<{ lat: number; lon: number }> {
  return new Promise((resolve, reject) => {
    if (typeof navigator === 'undefined' || !navigator.geolocation) {
      reject(new Error('unavailable'));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) =>
        resolve({ lat: roundCoordinate(pos.coords.latitude), lon: roundCoordinate(pos.coords.longitude) }),
      (err) => reject(new Error(err.message)),
      { timeout, maximumAge: 300_000 },
    );
  });
}
