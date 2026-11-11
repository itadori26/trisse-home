import type { PeerList } from '../domain/peer';

/** Output port: WireGuard backend. */
export interface WireGuardGateway {
  peers(): Promise<PeerList>;
  rename(publicKey: string, name: string): Promise<void>;
}

/** Application logic of the WireGuard feature. */
export class WireGuardUseCases {
  constructor(private readonly gateway: WireGuardGateway) {}

  peers(): Promise<PeerList> {
    return this.gateway.peers();
  }

  /** Sets a custom name; an empty name falls back to the exported one. */
  rename(publicKey: string, name: string): Promise<void> {
    return this.gateway.rename(publicKey, name.trim());
  }
}
