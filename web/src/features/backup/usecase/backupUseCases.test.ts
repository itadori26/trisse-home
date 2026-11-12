import { describe, expect, it, vi } from 'vitest';
import { BackupUseCases, type BackupGateway } from './backupUseCases';

describe('BackupUseCases', () => {
  it('sorts backups newest first', async () => {
    const gateway = {
      list: vi.fn().mockResolvedValue([
        { name: 'a', size: 1, createdAt: new Date(1) },
        { name: 'b', size: 1, createdAt: new Date(2) },
      ]),
      create: vi.fn(),
      remoteSettings: vi.fn(),
      saveRemoteSettings: vi.fn(),
      testRemote: vi.fn(),
      disableRemote: vi.fn(),
      listRemote: vi.fn(),
      restore: vi.fn(),
    } satisfies BackupGateway;
    expect((await new BackupUseCases(gateway).list()).map((f) => f.name)).toEqual(['b', 'a']);
  });
});
