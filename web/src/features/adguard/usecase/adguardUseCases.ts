import type { AdGuardSettings, AdGuardStats, DnsProvider, PauseMinutes } from '../domain/adguard';

/** Output port: DNS filter backend. */
export interface AdGuardGateway {
  settings(provider: DnsProvider): Promise<AdGuardSettings>;
  saveSettings(
    provider: DnsProvider,
    url: string,
    username: string,
    password: string,
  ): Promise<AdGuardSettings>;
  stats(provider: DnsProvider): Promise<AdGuardStats>;
  pause(provider: DnsProvider, minutes: PauseMinutes): Promise<AdGuardStats>;
}

/** Application logic of the DNS filter feature (AdGuard Home and Pi-hole). */
export class AdGuardUseCases {
  constructor(private readonly gateway: AdGuardGateway) {}

  settings(provider: DnsProvider = 'adguard'): Promise<AdGuardSettings> {
    return this.gateway.settings(provider);
  }

  /** Saves the connection; an empty password keeps the stored one. */
  saveSettings(
    provider: DnsProvider,
    url: string,
    username: string,
    password: string,
  ): Promise<AdGuardSettings> {
    return this.gateway.saveSettings(provider, url.trim(), username.trim(), password);
  }

  stats(provider: DnsProvider = 'adguard'): Promise<AdGuardStats> {
    return this.gateway.stats(provider);
  }

  /** Pauses blocking (#177); only allowed when the server enables it. */
  pause(provider: DnsProvider, minutes: PauseMinutes): Promise<AdGuardStats> {
    return this.gateway.pause(provider, minutes);
  }
}
