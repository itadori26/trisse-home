import { describe, expect, it, vi } from 'vitest';
import type { BackupSource } from '../domain/source';
import { BackupMonitorUseCases, type BackupMonitorGateway } from './backupMonitorUseCases';

const src = (name: string, state: BackupSource['state']): BackupSource => ({
  id: name,
  name,
  maxAgeHours: 24,
  state,
  lastSuccessAt: null,
  lastFailureAt: null,
  lastMessage: '',
});

describe('BackupMonitorUseCases', () => {
  it('lists problems first', async () => {
    const gateway = {
      list: vi.fn().mockResolvedValue([src('a', 'ok'), src('b', 'failed'), src('c', 'overdue')]),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      rotateToken: vi.fn(),
    } satisfies BackupMonitorGateway;
    expect((await new BackupMonitorUseCases(gateway).list()).map((s) => s.name)).toEqual(['b', 'c', 'a']);
  });
});
