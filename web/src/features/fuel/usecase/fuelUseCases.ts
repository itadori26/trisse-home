import {
  type Favourite,
  type FuelOverview,
  type FuelSettings,
  type FuelType,
  validateSettings,
} from '../domain/fuel';

/** Settings to save; an empty API key keeps the stored one. */
export interface FuelSettingsInput {
  apiKey: string;
  lat: number;
  lon: number;
  place: string;
  radiusKm: number;
  fuel: FuelType;
}

/** Output port: fuel backend. */
export interface FuelGateway {
  settings(): Promise<FuelSettings>;
  saveSettings(input: FuelSettingsInput): Promise<FuelSettings>;
  setFavourites(list: Favourite[]): Promise<FuelSettings>;
  overview(): Promise<FuelOverview>;
}

/** Application logic of the fuel price feature (#53). */
export class FuelUseCases {
  constructor(private readonly gateway: FuelGateway) {}

  settings(): Promise<FuelSettings> {
    return this.gateway.settings();
  }

  /** Validates, tests and saves the settings. */
  saveSettings(input: FuelSettingsInput): Promise<FuelSettings> {
    const problem = validateSettings(input.lat, input.lon, input.radiusKm);
    if (problem) return Promise.reject(new Error(problem));
    return this.gateway.saveSettings({ ...input, apiKey: input.apiKey.trim(), place: input.place.trim() });
  }

  setFavourites(list: Favourite[]): Promise<FuelSettings> {
    return this.gateway.setFavourites(list);
  }

  overview(): Promise<FuelOverview> {
    return this.gateway.overview();
  }
}
