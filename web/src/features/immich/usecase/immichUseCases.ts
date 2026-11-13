import type { ImmichOverview, ImmichSettings } from '../domain/immich';

/** Output port: Immich backend. */
export interface ImmichGateway {
  settings(): Promise<ImmichSettings>;
  saveSettings(url: string, apiKey: string): Promise<ImmichSettings>;
  overview(): Promise<ImmichOverview>;
}

/** Application logic of the Immich feature (#85). */
export class ImmichUseCases {
  constructor(private readonly gateway: ImmichGateway) {}

  settings(): Promise<ImmichSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; an empty key keeps the stored one. */
  saveSettings(url: string, apiKey: string): Promise<ImmichSettings> {
    return this.gateway.saveSettings(url.trim(), apiKey.trim());
  }

  overview(): Promise<ImmichOverview> {
    return this.gateway.overview();
  }
}
