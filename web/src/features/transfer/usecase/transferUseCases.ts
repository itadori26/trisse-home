import { type ImportResult, parseExportFile } from '../domain/transfer';

/** Output port: transfer backend. */
export interface TransferGateway {
  exportUrl(): string;
  import(file: unknown): Promise<ImportResult>;
}

/** Application logic of the transfer feature. */
export class TransferUseCases {
  constructor(private readonly gateway: TransferGateway) {}

  /** URL that downloads the export file. */
  exportUrl(): string {
    return this.gateway.exportUrl();
  }

  /** Validates the file locally and imports it. */
  async importFile(text: string): Promise<ImportResult> {
    const parsed = parseExportFile(text);
    if (typeof parsed === 'string') throw new Error(parsed);
    return this.gateway.import(parsed);
  }
}
