import type { ForecastSettings, HistoryPoint, SystemSnapshot, VersionInfo } from '../domain/system';

/** Output port: system metrics backend. */
export interface SystemGateway {
  snapshot(): Promise<SystemSnapshot>;
  history(hours: number): Promise<HistoryPoint[]>;
  version(): Promise<VersionInfo>;
  setUpdateCheck(enabled: boolean): Promise<VersionInfo>;
  forecastSettings(): Promise<ForecastSettings>;
  saveForecastSettings(s: ForecastSettings): Promise<ForecastSettings>;
}

/** Application logic of the system feature. */
export class SystemUseCases {
  constructor(private readonly gateway: SystemGateway) {}

  snapshot(): Promise<SystemSnapshot> {
    return this.gateway.snapshot();
  }

  /** History of the last 24 hours. */
  history(): Promise<HistoryPoint[]> {
    return this.gateway.history(24);
  }

  version(): Promise<VersionInfo> {
    return this.gateway.version();
  }

  setUpdateCheck(enabled: boolean): Promise<VersionInfo> {
    return this.gateway.setUpdateCheck(enabled);
  }

  forecastSettings(): Promise<ForecastSettings> {
    return this.gateway.forecastSettings();
  }

  saveForecastSettings(s: ForecastSettings): Promise<ForecastSettings> {
    return this.gateway.saveForecastSettings({
      windowDays: Math.round(s.windowDays),
      alertDays: Math.round(s.alertDays),
    });
  }
}
