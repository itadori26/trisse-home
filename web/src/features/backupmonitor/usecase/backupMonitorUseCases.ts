import type { BackupSource, CreatedSource, SourceSchedule } from '../domain/source';

/** Output port: backup monitor backend. */
export interface BackupMonitorGateway {
  list(): Promise<BackupSource[]>;
  create(name: string, schedule: SourceSchedule): Promise<CreatedSource>;
  update(id: string, name: string, schedule: SourceSchedule): Promise<BackupSource>;
  delete(id: string): Promise<void>;
  rotateToken(id: string): Promise<CreatedSource>;
}

/** Application logic of the backup monitor feature. */
export class BackupMonitorUseCases {
  constructor(private readonly gateway: BackupMonitorGateway) {}

  /** Lists sources; problems first. */
  async list(): Promise<BackupSource[]> {
    const rank = { failed: 0, overdue: 1, running: 2, waiting: 3, ok: 4 } as const;
    return [...(await this.gateway.list())].sort(
      (a, b) => rank[a.state] - rank[b.state] || a.name.localeCompare(b.name),
    );
  }

  create(name: string, schedule: SourceSchedule): Promise<CreatedSource> {
    return this.gateway.create(name.trim(), { ...schedule, cron: schedule.cron.trim() });
  }

  update(id: string, name: string, schedule: SourceSchedule): Promise<BackupSource> {
    return this.gateway.update(id, name.trim(), { ...schedule, cron: schedule.cron.trim() });
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  rotateToken(id: string): Promise<CreatedSource> {
    return this.gateway.rotateToken(id);
  }
}
