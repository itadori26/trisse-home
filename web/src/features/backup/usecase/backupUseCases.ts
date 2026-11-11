import type { BackupFile } from '../domain/backup';

/** Output port: backup backend. */
export interface BackupGateway {
  list(): Promise<BackupFile[]>;
  create(): Promise<BackupFile>;
}

/** Application logic of the backup feature. */
export class BackupUseCases {
  constructor(private readonly gateway: BackupGateway) {}

  /** Lists backups, newest first. */
  async list(): Promise<BackupFile[]> {
    return [...(await this.gateway.list())].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
  }

  backupNow(): Promise<BackupFile> {
    return this.gateway.create();
  }
}
