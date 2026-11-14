import { type DropItem, type ExpiryHours, validateFile } from '../domain/drop';

/** Output port: drop zone backend. */
export interface DropGateway {
  list(): Promise<DropItem[]>;
  addText(text: string, expiresInHours: ExpiryHours): Promise<DropItem>;
  addFile(file: File, expiresInHours: ExpiryHours): Promise<DropItem>;
  remove(id: string): Promise<void>;
  fileUrl(id: string): string;
}

/** Application logic of the drop zone (#101). */
export class DropUseCases {
  constructor(private readonly gateway: DropGateway) {}

  list(): Promise<DropItem[]> {
    return this.gateway.list();
  }

  addText(text: string, expiresInHours: ExpiryHours): Promise<DropItem> {
    return this.gateway.addText(text.trim(), expiresInHours);
  }

  /** Uploads files one after another; stops at the first that is refused. */
  async addFiles(files: File[], expiresInHours: ExpiryHours): Promise<DropItem[]> {
    const out: DropItem[] = [];
    for (const f of files) {
      const problem = validateFile(f);
      if (problem) throw new Error(`${f.name}: ${problem}`);
      out.push(await this.gateway.addFile(f, expiresInHours));
    }
    return out;
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  fileUrl(id: string): string {
    return this.gateway.fileUrl(id);
  }
}
