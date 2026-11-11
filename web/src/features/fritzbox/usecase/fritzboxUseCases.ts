import type { FritzBoxSettings, FritzBoxStatus } from '../domain/fritzbox';

/** Output port: FRITZ!Box backend. */
export interface FritzBoxGateway {
  settings(): Promise<FritzBoxSettings>;
  saveSettings(url: string, username: string, password: string): Promise<FritzBoxSettings>;
  status(): Promise<FritzBoxStatus>;
}

/** Application logic of the FRITZ!Box feature. */
export class FritzBoxUseCases {
  constructor(private readonly gateway: FritzBoxGateway) {}

  settings(): Promise<FritzBoxSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; user and password are optional. */
  saveSettings(url: string, username: string, password: string): Promise<FritzBoxSettings> {
    return this.gateway.saveSettings(url.trim(), username.trim(), password);
  }

  status(): Promise<FritzBoxStatus> {
    return this.gateway.status();
  }
}
