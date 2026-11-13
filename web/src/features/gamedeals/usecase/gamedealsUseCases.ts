import { type GameDealsOverview, type GameDealsSettings, steamIdFrom } from '../domain/gamedeals';

/** Output port: game deals backend. */
export interface GameDealsGateway {
  settings(): Promise<GameDealsSettings>;
  saveSettings(s: GameDealsSettings): Promise<GameDealsSettings>;
  overview(): Promise<GameDealsOverview>;
}

/** Application logic of the game deals feature (#62). */
export class GamedealsUseCases {
  constructor(private readonly gateway: GameDealsGateway) {}

  settings(): Promise<GameDealsSettings> {
    return this.gateway.settings();
  }

  /** Saves the settings; a pasted profile URL is reduced to the Steam ID. */
  saveSettings(s: GameDealsSettings): Promise<GameDealsSettings> {
    return this.gateway.saveSettings({
      ...s,
      steamId: steamIdFrom(s.steamId),
      country: s.country.trim().toUpperCase(),
    });
  }

  overview(): Promise<GameDealsOverview> {
    return this.gateway.overview();
  }
}
