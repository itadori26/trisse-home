import type { AdGuardSettings, AdGuardStats } from '../domain/adguard';

/** Output port: AdGuard backend. */
export interface AdGuardGateway {
  settings(): Promise<AdGuardSettings>;
  saveSettings(url: string, username: string, password: string): Promise<AdGuardSettings>;
  stats(): Promise<AdGuardStats>;
}

/** Application logic of the AdGuard feature. */
export class AdGuardUseCases {
  constructor(private readonly gateway: AdGuardGateway) {}

  settings(): Promise<AdGuardSettings> {
    return this.gateway.settings();
  }

  /** Saves the connection; an empty password keeps the stored one. */
  saveSettings(url: string, username: string, password: string): Promise<AdGuardSettings> {
    return this.gateway.saveSettings(url.trim(), username.trim(), password);
  }

  stats(): Promise<AdGuardStats> {
    return this.gateway.stats();
  }
}
