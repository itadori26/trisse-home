import type { HttpClient } from '../../../../shared/api-client/http';
import type { FileDTO } from '../../../../shared/api-client/generated/backup';
import type { BackupFile } from '../../domain/backup';
import type { BackupGateway } from '../../usecase/backupUseCases';

const toFile = (d: FileDTO): BackupFile => ({ name: d.name, size: d.size, createdAt: new Date(d.createdAt) });

/** BackupGateway backed by the REST API. */
export class BackupApi implements BackupGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<BackupFile[]> {
    return (await this.http.get<FileDTO[]>('/api/backups')).map(toFile);
  }

  async create(): Promise<BackupFile> {
    return toFile(await this.http.post<FileDTO>('/api/backups'));
  }
}
