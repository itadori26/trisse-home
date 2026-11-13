import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, SettingsDTO } from '../../../../shared/api-client/generated/fuel';
import type { FavouriteDTO, SettingsInputDTO } from '../../../../shared/api-client/generated/fuel-requests';
import type { Favourite, FuelOverview, FuelSettings } from '../../domain/fuel';
import type { FuelGateway, FuelSettingsInput } from '../../usecase/fuelUseCases';

const toOverview = (d: OverviewDTO): FuelOverview => ({
  fuel: d.fuel,
  place: d.place,
  cheapest: d.cheapest.map((s) => ({ ...s, price: s.price ?? null })),
  favourites: d.favourites.map((s) => ({ ...s, price: s.price ?? null })),
});

/** FuelGateway backed by the REST API. */
export class FuelApi implements FuelGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<FuelSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/fuel/settings')) };
  }
  async saveSettings(input: FuelSettingsInput): Promise<FuelSettings> {
    const body: SettingsInputDTO = input;
    return { ...(await this.http.put<SettingsDTO>('/api/fuel/settings', body)) };
  }
  async setFavourites(list: Favourite[]): Promise<FuelSettings> {
    const body: FavouriteDTO[] = list;
    return { ...(await this.http.put<SettingsDTO>('/api/fuel/favourites', body)) };
  }
  async overview(): Promise<FuelOverview> {
    return toOverview(await this.http.get<OverviewDTO>('/api/fuel/overview'));
  }
}
