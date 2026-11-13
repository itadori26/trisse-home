import type { Entity, SmartHomeSettings } from '../domain/smarthome';

/** Output port: smart home backend. */
export interface SmartHomeGateway {
  settings(): Promise<SmartHomeSettings>;
  saveSettings(url: string, token: string): Promise<SmartHomeSettings>;
  entities(): Promise<Entity[]>;
  states(ids: string[]): Promise<Entity[]>;
  act(id: string): Promise<void>;
}

/** Application logic of the smart home feature (#49). */
export class SmarthomeUseCases {
  constructor(private readonly gateway: SmartHomeGateway) {}

  settings(): Promise<SmartHomeSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; an empty token keeps the stored one. */
  saveSettings(url: string, token: string): Promise<SmartHomeSettings> {
    return this.gateway.saveSettings(url.trim(), token.trim());
  }

  /** Every entity, for the widget's entity picker. */
  entities(): Promise<Entity[]> {
    return this.gateway.entities();
  }

  /** The widget's entities in their configured order; none means no request. */
  states(ids: string[]): Promise<Entity[]> {
    return ids.length === 0 ? Promise.resolve([]) : this.gateway.states(ids);
  }

  /** Toggles or presses an entity; the UI asks for confirmation first. */
  act(id: string): Promise<void> {
    return this.gateway.act(id);
  }
}
