import type { HttpClient } from '../../../../shared/api-client/http';
import type { ListDTO } from '../../../../shared/api-client/generated/wireguard';
import type { RenameDTO } from '../../../../shared/api-client/generated/wireguard-requests';
import type { PeerList } from '../../domain/peer';
import type { WireGuardGateway } from '../../usecase/wireguardUseCases';

/** WireGuardGateway backed by the REST API. */
export class WireGuardApi implements WireGuardGateway {
  constructor(private readonly http: HttpClient) {}

  async peers(): Promise<PeerList> {
    const d = await this.http.get<ListDTO>('/api/wireguard/peers');
    return {
      peers: d.peers.map((p) => ({
        ...p,
        latestHandshake: p.latestHandshake ? new Date(p.latestHandshake) : null,
      })),
      online: d.online,
      updatedAt: new Date(d.updatedAt),
      stale: d.stale,
    };
  }

  rename(publicKey: string, name: string): Promise<void> {
    const body: RenameDTO = { publicKey, name };
    return this.http.put('/api/wireguard/names', body);
  }
}
