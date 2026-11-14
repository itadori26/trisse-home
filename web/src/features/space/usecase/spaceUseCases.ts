import type { Launch, Picture, SpaceSettings } from '../domain/space';

/** Output port: space backend. */
export interface SpaceGateway {
  launches(limit: number): Promise<Launch[]>;
  picture(): Promise<Picture>;
  settings(): Promise<SpaceSettings>;
  saveSettings(launches: boolean, picture: boolean, nasaKey: string): Promise<SpaceSettings>;
}

/** Application logic of the space feature (#64). */
export class SpaceUseCases {
  constructor(private readonly gateway: SpaceGateway) {}

  launches(limit: number): Promise<Launch[]> {
    return this.gateway.launches(limit);
  }

  picture(): Promise<Picture> {
    return this.gateway.picture();
  }

  settings(): Promise<SpaceSettings> {
    return this.gateway.settings();
  }

  /** Saves the switches; an empty key keeps the stored one. */
  saveSettings(launches: boolean, picture: boolean, nasaKey: string): Promise<SpaceSettings> {
    return this.gateway.saveSettings(launches, picture, nasaKey.trim());
  }
}
