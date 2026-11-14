import type { HolidaysOverview, State } from '../domain/holidays';

/** Output port: holidays backend. */
export interface HolidaysGateway {
  upcoming(state: State, limit: number): Promise<HolidaysOverview>;
  settings(): Promise<{ schoolUrl: string }>;
  saveSettings(schoolUrl: string): Promise<{ schoolUrl: string }>;
}

/** Application logic of the holidays feature (#55). */
export class HolidaysUseCases {
  constructor(private readonly gateway: HolidaysGateway) {}

  upcoming(state: State, limit: number): Promise<HolidaysOverview> {
    return this.gateway.upcoming(state, limit);
  }

  settings(): Promise<{ schoolUrl: string }> {
    return this.gateway.settings();
  }

  /** Saves (and tests) the ICS link; an empty link removes the school holidays. */
  saveSettings(schoolUrl: string): Promise<{ schoolUrl: string }> {
    return this.gateway.saveSettings(schoolUrl.trim());
  }
}
