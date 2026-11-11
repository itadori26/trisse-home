import type { NextcloudSettings, NextcloudStatus } from '../domain/nextcloud';

/** Output port: Nextcloud backend. */
export interface NextcloudGateway {
  settings(): Promise<NextcloudSettings>;
  saveSettings(url: string, username: string, appPassword: string): Promise<NextcloudSettings>;
  status(): Promise<NextcloudStatus>;
}

/** Application logic of the Nextcloud feature. */
export class NextcloudUseCases {
  constructor(private readonly gateway: NextcloudGateway) {}

  settings(): Promise<NextcloudSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; an empty app password keeps the stored one. */
  saveSettings(url: string, username: string, appPassword: string): Promise<NextcloudSettings> {
    return this.gateway.saveSettings(url.trim(), username.trim(), appPassword.trim());
  }

  status(): Promise<NextcloudStatus> {
    return this.gateway.status();
  }
}
