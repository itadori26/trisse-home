import type { Forecast, Place, WeatherConfig } from '../domain/weather';

/** Output port: weather backend. */
export interface WeatherGateway {
  forecast(lat: number, lon: number): Promise<Forecast>;
  searchPlaces(query: string): Promise<Place[]>;
}

/** Minimum length of a place search. */
export const MIN_PLACE_QUERY = 2;

/** Application logic of the weather feature. */
export class WeatherUseCases {
  constructor(private readonly gateway: WeatherGateway) {}

  /** Loads the forecast for a configured widget and trims it to the configured days. */
  async forecast(cfg: WeatherConfig): Promise<Forecast> {
    const f = await this.gateway.forecast(cfg.lat, cfg.lon);
    return { ...f, days: f.days.slice(0, cfg.days) };
  }

  /** Searches places; short queries return no results without a request. */
  searchPlaces(query: string): Promise<Place[]> {
    const q = query.trim();
    if (q.length < MIN_PLACE_QUERY) return Promise.resolve([]);
    return this.gateway.searchPlaces(q);
  }
}
