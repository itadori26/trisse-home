import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  MeasurementDTO,
  ResultsDTO,
  SettingsDTO,
} from '../../../../shared/api-client/generated/speedtest';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/speedtest-requests';
import type { Measurement, Range, SpeedResults, SpeedtestSettings } from '../../domain/speedtest';
import type { SpeedtestGateway, SpeedtestSettingsInput } from '../../usecase/speedtestUseCases';

const toMeasurement = (d: MeasurementDTO): Measurement => ({ ...d, at: new Date(d.at) });

/** SpeedtestGateway backed by the REST API. */
export class SpeedtestApi implements SpeedtestGateway {
  constructor(private readonly http: HttpClient) {}

  async results(hours: Range): Promise<SpeedResults> {
    const d = await this.http.get<ResultsDTO>(`/api/speedtest/results?hours=${hours}`);
    return {
      latest: d.latest ? toMeasurement(d.latest) : null,
      history: d.history.map(toMeasurement),
      error: d.error,
    };
  }

  settings(): Promise<SpeedtestSettings> {
    return this.http.get<SettingsDTO>('/api/speedtest/settings');
  }

  saveSettings(input: SpeedtestSettingsInput): Promise<SpeedtestSettings> {
    const body: SettingsInputDTO = input;
    return this.http.put<SettingsDTO>('/api/speedtest/settings', body);
  }
}
