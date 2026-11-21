import type { HttpClient } from '../../../../shared/api-client/http';
import type { FileDTO, RemoteSettingsDTO, StatusDTO } from '../../../../shared/api-client/generated/backup';
import type { RemoteSettingsInputDTO } from '../../../../shared/api-client/generated/backup-requests';
import {
  type BackupFile,
  type BackupHealth,
  type BackupStatus,
  type RemoteSettings,
  type RemoteSettingsInput,
  type RestoreRequest,
  restoreForm,
} from '../../domain/backup';
import type { BackupGateway } from '../../usecase/backupUseCases';

const toFile = (d: FileDTO): BackupFile => ({ name: d.name, size: d.size, createdAt: new Date(d.createdAt) });
const toDate = (s: string | null | undefined) => (s ? new Date(s) : null);

/** BackupGateway backed by the REST API. */
export class BackupApi implements BackupGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<BackupFile[]> {
    return (await this.http.get<FileDTO[]>('/api/backups')).map(toFile);
  }

  async create(): Promise<BackupFile> {
    return toFile(await this.http.post<FileDTO>('/api/backups'));
  }

  remoteSettings(): Promise<RemoteSettings> {
    return this.http.get<RemoteSettingsDTO>('/api/backups/remote/settings');
  }

  saveRemoteSettings(input: RemoteSettingsInput): Promise<RemoteSettings> {
    const body: RemoteSettingsInputDTO = input;
    return this.http.put<RemoteSettingsDTO>('/api/backups/remote/settings', body);
  }

  async testRemote(input: RemoteSettingsInput): Promise<void> {
    const body: RemoteSettingsInputDTO = input;
    await this.http.post<unknown>('/api/backups/remote/test', body);
  }

  disableRemote(): Promise<void> {
    return this.http.delete('/api/backups/remote/settings');
  }

  async listRemote(): Promise<BackupFile[]> {
    return (await this.http.get<FileDTO[]>('/api/backups/remote')).map(toFile);
  }

  async status(): Promise<BackupStatus> {
    const d = await this.http.get<StatusDTO>('/api/backups/status');
    return {
      health: d.health as BackupHealth,
      lastBackupAt: toDate(d.lastBackupAt),
      count: d.count,
      failedAt: toDate(d.failedAt),
      error: d.error,
      offsiteFailedAt: toDate(d.offsiteFailedAt),
      offsiteError: d.offsiteError,
    };
  }

  async restore(req: RestoreRequest): Promise<void> {
    await this.http.postForm<unknown>('/api/backups/restore', restoreForm(req));
  }
}
