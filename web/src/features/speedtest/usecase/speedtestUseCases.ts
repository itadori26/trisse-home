import type { Range, SpeedResults, SpeedtestSettings } from '../domain/speedtest';

/** Input of saveSettings; an empty token keeps the stored one. */
export interface SpeedtestSettingsInput {
  url: string;
  token: string;
  minDownMbps: number;
  minUpMbps: number;
}

/** Output port: speedtest backend. */
export interface SpeedtestGateway {
  results(hours: Range): Promise<SpeedResults>;
  settings(): Promise<SpeedtestSettings>;
  saveSettings(input: SpeedtestSettingsInput): Promise<SpeedtestSettings>;
}

/** Application logic of the speedtest feature. */
export class SpeedtestUseCases {
  constructor(private readonly gateway: SpeedtestGateway) {}

  results(hours: Range): Promise<SpeedResults> {
    return this.gateway.results(hours);
  }

  settings(): Promise<SpeedtestSettings> {
    return this.gateway.settings();
  }

  saveSettings(input: SpeedtestSettingsInput): Promise<SpeedtestSettings> {
    return this.gateway.saveSettings({
      ...input,
      url: input.url.trim(),
      minDownMbps: Math.max(0, input.minDownMbps || 0),
      minUpMbps: Math.max(0, input.minUpMbps || 0),
    });
  }
}
