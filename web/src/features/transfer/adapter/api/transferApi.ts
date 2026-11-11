import type { HttpClient } from '../../../../shared/api-client/http';
import type { ImportResultDTO } from '../../../../shared/api-client/generated/transfer';
import type { ImportResult } from '../../domain/transfer';
import type { TransferGateway } from '../../usecase/transferUseCases';

/** TransferGateway backed by the REST API. */
export class TransferApi implements TransferGateway {
  constructor(private readonly http: HttpClient) {}

  exportUrl(): string {
    return '/api/transfer/export';
  }

  async import(file: unknown): Promise<ImportResult> {
    const d = await this.http.post<ImportResultDTO>('/api/transfer/import', file);
    return { imported: d.imported ?? [], skipped: d.skipped ?? [], failed: d.failed ?? {} };
  }
}
