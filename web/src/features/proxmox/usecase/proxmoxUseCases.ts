import type { GuestAction, ProxmoxSettings, ProxmoxStatus } from '../domain/proxmox';

/** Settings to save; an empty secret keeps the stored one. */
export interface ProxmoxSettingsInput {
  url: string;
  tokenId: string;
  secret: string;
  fingerprint: string;
}

/** Output port: Proxmox backend. */
export interface ProxmoxGateway {
  settings(): Promise<ProxmoxSettings>;
  saveSettings(input: ProxmoxSettingsInput): Promise<ProxmoxSettings>;
  status(): Promise<ProxmoxStatus>;
  act(guestId: number, action: GuestAction): Promise<void>;
}

/** Application logic of the Proxmox feature (#84). */
export class ProxmoxUseCases {
  constructor(private readonly gateway: ProxmoxGateway) {}

  settings(): Promise<ProxmoxSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; values are trimmed, the secret is not. */
  saveSettings(input: ProxmoxSettingsInput): Promise<ProxmoxSettings> {
    return this.gateway.saveSettings({
      url: input.url.trim(),
      tokenId: input.tokenId.trim(),
      secret: input.secret,
      fingerprint: input.fingerprint.trim(),
    });
  }

  status(): Promise<ProxmoxStatus> {
    return this.gateway.status();
  }

  /** Starts or shuts down a guest; the UI asks for confirmation first. */
  act(guestId: number, action: GuestAction): Promise<void> {
    return this.gateway.act(guestId, action);
  }
}
