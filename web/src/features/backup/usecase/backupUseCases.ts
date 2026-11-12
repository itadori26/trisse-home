import type { BackupFile, RemoteSettings, RemoteSettingsInput, RestoreRequest } from '../domain/backup';

/** Output port: backup backend. */
export interface BackupGateway {
  list(): Promise<BackupFile[]>;
  create(): Promise<BackupFile>;
  remoteSettings(): Promise<RemoteSettings>;
  saveRemoteSettings(input: RemoteSettingsInput): Promise<RemoteSettings>;
  testRemote(input: RemoteSettingsInput): Promise<void>;
  disableRemote(): Promise<void>;
  listRemote(): Promise<BackupFile[]>;
  restore(req: RestoreRequest): Promise<void>;
}

const newestFirst = (files: BackupFile[]) =>
  [...files].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());

/** Application logic of the backup feature. */
export class BackupUseCases {
  constructor(private readonly gateway: BackupGateway) {}

  /** Lists backups, newest first. */
  async list(): Promise<BackupFile[]> {
    return newestFirst(await this.gateway.list());
  }

  backupNow(): Promise<BackupFile> {
    return this.gateway.create();
  }

  /** Off-site target (#46). */
  remoteSettings(): Promise<RemoteSettings> {
    return this.gateway.remoteSettings();
  }

  saveRemoteSettings(input: RemoteSettingsInput): Promise<RemoteSettings> {
    return this.gateway.saveRemoteSettings({
      ...input,
      url: input.url.trim(),
      username: input.username.trim(),
    });
  }

  testRemote(input: RemoteSettingsInput): Promise<void> {
    return this.gateway.testRemote({ ...input, url: input.url.trim(), username: input.username.trim() });
  }

  disableRemote(): Promise<void> {
    return this.gateway.disableRemote();
  }

  async listRemote(): Promise<BackupFile[]> {
    return newestFirst(await this.gateway.listRemote());
  }

  /** Restores a backup; the server restarts afterwards. */
  restore(req: RestoreRequest): Promise<void> {
    return this.gateway.restore(req);
  }
}
