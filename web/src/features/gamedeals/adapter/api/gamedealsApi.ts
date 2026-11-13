import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, SettingsDTO } from '../../../../shared/api-client/generated/gamedeals';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/gamedeals-requests';
import type { GameDealsOverview, GameDealsSettings } from '../../domain/gamedeals';
import type { GameDealsGateway } from '../../usecase/gamedealsUseCases';

/** GameDealsGateway backed by the REST API. */
export class GamedealsApi implements GameDealsGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<GameDealsSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/gamedeals/settings')) };
  }
  async saveSettings(s: GameDealsSettings): Promise<GameDealsSettings> {
    const body: SettingsInputDTO = s;
    return { ...(await this.http.put<SettingsDTO>('/api/gamedeals/settings', body)) };
  }
  async overview(): Promise<GameDealsOverview> {
    const d = await this.http.get<OverviewDTO>('/api/gamedeals/overview');
    const game = (g: OverviewDTO['freeNow'][number]) => ({
      ...g,
      start: new Date(g.start),
      end: new Date(g.end),
    });
    return {
      deals: d.deals.map((x) => ({ ...x, endsAt: x.endsAt ? new Date(x.endsAt) : null })),
      freeNow: d.freeNow.map(game),
      freeNext: d.freeNext.map(game),
      errors: d.errors,
    };
  }
}
