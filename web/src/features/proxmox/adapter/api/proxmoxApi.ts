import type { HttpClient } from '../../../../shared/api-client/http';
import type { SettingsDTO, StatusDTO } from '../../../../shared/api-client/generated/proxmox';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/proxmox-requests';
import type { GuestAction, ProxmoxSettings, ProxmoxStatus } from '../../domain/proxmox';
import type { ProxmoxGateway, ProxmoxSettingsInput } from '../../usecase/proxmoxUseCases';

/** ProxmoxGateway backed by the REST API. */
export class ProxmoxApi implements ProxmoxGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<ProxmoxSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/proxmox/settings')) };
  }
  async saveSettings(input: ProxmoxSettingsInput): Promise<ProxmoxSettings> {
    const body: SettingsInputDTO = input;
    return { ...(await this.http.put<SettingsDTO>('/api/proxmox/settings', body)) };
  }
  async status(): Promise<ProxmoxStatus> {
    const d = await this.http.get<StatusDTO>('/api/proxmox/status');
    return { nodes: d.nodes, guests: d.guests, error: d.error };
  }
  act(guestId: number, action: GuestAction): Promise<void> {
    return this.http.post(`/api/proxmox/guests/${guestId}/${action}`);
  }
}
