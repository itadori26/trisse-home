import type { HttpClient } from '../../../../shared/api-client/http';
import type { StateDTO } from '../../../../shared/api-client/generated/device';
import type { ModeDTO } from '../../../../shared/api-client/generated/device-requests';
import type { DeviceState } from '../../domain/device';
import type { DeviceGateway } from '../../usecase/deviceUseCases';

/** DeviceGateway backed by the REST API. */
export class DeviceApi implements DeviceGateway {
  constructor(private readonly http: HttpClient) {}

  async state(): Promise<DeviceState> {
    return { ...(await this.http.get<StateDTO>('/api/device')) };
  }
  async changeMode(mode: 'headless' | 'kiosk'): Promise<DeviceState> {
    const body: ModeDTO = { mode };
    return { ...(await this.http.put<StateDTO>('/api/device/display-mode', body)) };
  }
}
