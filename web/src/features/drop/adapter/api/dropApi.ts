import type { HttpClient } from '../../../../shared/api-client/http';
import type { ItemDTO } from '../../../../shared/api-client/generated/drop';
import type { DropItem, ExpiryHours } from '../../domain/drop';
import type { DropGateway } from '../../usecase/dropUseCases';

const toItem = (d: ItemDTO): DropItem => ({
  ...d,
  createdAt: new Date(d.createdAt),
  expiresAt: new Date(d.expiresAt),
});

/** DropGateway backed by the REST API. */
export class DropApi implements DropGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<DropItem[]> {
    return (await this.http.get<ItemDTO[]>('/api/drop')).map(toItem);
  }
  async addText(text: string, expiresInHours: ExpiryHours): Promise<DropItem> {
    return toItem(await this.http.post<ItemDTO>('/api/drop/text', { text, expiresInHours }));
  }
  async addFile(file: File, expiresInHours: ExpiryHours): Promise<DropItem> {
    const form = new FormData();
    form.append('expiresInHours', String(expiresInHours));
    form.append('file', file, file.name);
    return toItem(await this.http.postForm<ItemDTO>('/api/drop/files', form));
  }
  remove(id: string): Promise<void> {
    return this.http.delete(`/api/drop/${encodeURIComponent(id)}`);
  }
  fileUrl(id: string): string {
    return `/api/drop/${encodeURIComponent(id)}/file`;
  }
}
