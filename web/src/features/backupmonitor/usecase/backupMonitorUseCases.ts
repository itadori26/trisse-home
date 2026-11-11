import type { BackupSource, CreatedSource } from '../domain/source';

/** Output port: backup monitor backend. */
export interface BackupMonitorGateway {
  list(): Promise<BackupSource[]>;
  create(name: string, maxAgeHours: number): Promise<CreatedSource>;
  update(id: string, name: string, maxAgeHours: number): Promise<BackupSource>;
  delete(id: string): Promise<void>;
  rotateToken(id: string): Promise<CreatedSource>;
}

/** Application logic of the backup monitor feature. */
export class BackupMonitorUseCases {
  constructor(private readonly gateway: BackupMonitorGateway) {}

  /** Lists sources; problems first. */
  async list(): Promise<BackupSource[]> {
    const rank = { failed: 0, overdue: 1, waiting: 2, ok: 3 } as const;
    return [...(await this.gateway.list())].sort(
      (a, b) => rank[a.state] - rank[b.state] || a.name.localeCompare(b.name),
    );
  }

  create(name: string, maxAgeHours: number): Promise<CreatedSource> {
    return this.gateway.create(name.trim(), maxAgeHours);
  }

  update(id: string, name: string, maxAgeHours: number): Promise<BackupSource> {
    return this.gateway.update(id, name.trim(), maxAgeHours);
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  rotateToken(id: string): Promise<CreatedSource> {
    return this.gateway.rotateToken(id);
  }
}
