import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  ForecastSettingsDTO,
  PointDTO,
  SnapshotDTO,
  VersionDTO,
} from '../../../../shared/api-client/generated/system';
import type {
  ForecastSettingsDTO as ForecastSettingsInputDTO,
  UpdateCheckDTO,
} from '../../../../shared/api-client/generated/system-requests';
import {
  type ForecastSettings,
  type HistoryPoint,
  parseInstallMethod,
  type SystemSnapshot,
  type VersionInfo,
} from '../../domain/system';
import type { SystemGateway } from '../../usecase/systemUseCases';

/** SystemGateway backed by the REST API. */
export class SystemApi implements SystemGateway {
  constructor(private readonly http: HttpClient) {}

  async snapshot(): Promise<SystemSnapshot> {
    const d = await this.http.get<SnapshotDTO>('/api/system');
    return {
      cpuPercent: d.cpuPercent,
      cpuCount: d.cpuCount,
      load: [d.load[0] ?? 0, d.load[1] ?? 0, d.load[2] ?? 0],
      memory: d.memory,
      swap: d.swap,
      disks: d.disks.map((x) => ({ label: x.label, usage: x.usage, forecast: x.forecast ?? null })),
      uptimeSeconds: d.uptimeSeconds,
      temperature: d.temperature,
    };
  }

  async history(hours: number): Promise<HistoryPoint[]> {
    const dtos = await this.http.get<PointDTO[]>(`/api/system/history?hours=${hours}`);
    return dtos.map((d) => ({ at: new Date(d.at), cpu: d.cpu, memory: d.memory, disk: d.disk }));
  }

  async version(): Promise<VersionInfo> {
    return toVersion(await this.http.get<VersionDTO>('/api/system/version'));
  }

  async setUpdateCheck(enabled: boolean): Promise<VersionInfo> {
    const body: UpdateCheckDTO = { enabled };
    return toVersion(await this.http.put<VersionDTO>('/api/system/update-check', body));
  }

  forecastSettings(): Promise<ForecastSettings> {
    return this.http.get<ForecastSettingsDTO>('/api/system/forecast-settings');
  }

  saveForecastSettings(s: ForecastSettings): Promise<ForecastSettings> {
    const body: ForecastSettingsInputDTO = s;
    return this.http.put<ForecastSettingsDTO>('/api/system/forecast-settings', body);
  }
}

const toVersion = (d: VersionDTO): VersionInfo => ({
  current: d.current,
  installMethod: parseInstallMethod(d.installMethod),
  checkEnabled: d.checkEnabled,
  latest: d.latest ? { ...d.latest, publishedAt: new Date(d.latest.publishedAt) } : null,
  updateAvailable: d.updateAvailable,
});
