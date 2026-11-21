import { describe, expect, it } from 'vitest';
import type { BackupStatus } from '../features/backup';
import { loadAreaStatuses } from './settingsStatus';
import { loadRecent, rememberChange } from './settingsRecent';

const now = new Date(2026, 10, 21, 12, 0);

const status = (s: Partial<BackupStatus>): BackupStatus => ({
  health: 'ok',
  lastBackupAt: null,
  count: s.lastBackupAt ? 1 : 0,
  failedAt: null,
  error: '',
  offsiteFailedAt: null,
  offsiteError: '',
  ...s,
});

describe('area statuses on the settings start page (#600)', () => {
  it('reads each status from its source', async () => {
    const s = await loadAreaStatuses(
      ['account', 'notifications', 'connections', 'data', 'system'],
      {
        twoFactor: async () => ({ enabled: false }),
        channels: async () => [{ enabled: true }, { enabled: false }, { enabled: true }],
        backups: async () => status({ lastBackupAt: new Date(2026, 10, 21, 3, 0) }),
        version: async () => ({ current: '0.2.0', updateAvailable: true, latest: { version: '0.3.0' } }),
        connections: 4,
      },
      now,
    );
    expect(s.account).toEqual({ text: 'Two-factor off', tone: 'warn' });
    expect(s.notifications).toEqual({ text: '2 channels on', tone: 'ok' });
    expect(s.connections).toEqual({ text: '4 integrations', tone: 'info' });
    expect(s.data?.tone).toBe('ok');
    expect(s.data?.text).toMatch(/^Last backup today 03:00.*, all good$/);
    expect(s.system).toEqual({ text: 'Update to 0.3.0 available', tone: 'warn' });
  });

  it('shows the backup traffic light: amber for a missing or old backup, red for a failure (#677)', async () => {
    const data = async (s: BackupStatus) =>
      (await loadAreaStatuses(['data'], { backups: async () => s }, now)).data;
    expect(await data(status({ health: 'stale' }))).toEqual({ text: 'No backup yet', tone: 'warn' });
    expect(await data(status({ health: 'stale', lastBackupAt: new Date(2026, 10, 20, 3, 0) }))).toEqual({
      text: expect.stringMatching(/^Last backup yesterday 03:00.*, older than expected$/),
      tone: 'warn',
    });
    expect(
      await data(
        status({
          health: 'stale',
          lastBackupAt: new Date(2026, 10, 21, 3, 0),
          offsiteFailedAt: new Date(2026, 10, 21, 3, 1),
        }),
      ),
    ).toEqual({
      text: expect.stringMatching(/^Last backup today .*, but the off-site copy failed$/),
      tone: 'warn',
    });
    const failed = await data(
      status({
        health: 'failed',
        lastBackupAt: new Date(2026, 10, 10, 3, 0),
        failedAt: new Date(2026, 10, 21, 3, 0),
      }),
    );
    expect(failed).toEqual({
      text: expect.stringMatching(/^The last backup failed \(today 03:00/),
      tone: 'error',
    });
  });

  it('leaves a tile without status when its source is missing or fails', async () => {
    const s = await loadAreaStatuses(
      ['account', 'system', 'people'],
      {
        twoFactor: async () => {
          throw new Error('403');
        },
      },
      now,
    );
    expect(s).toEqual({});
  });
});

describe('recently changed sections (#600)', () => {
  const memory = () => {
    const data = new Map<string, string>();
    return {
      getItem: (k: string) => data.get(k) ?? null,
      setItem: (k: string, v: string) => void data.set(k, v),
    };
  };

  it('remembers the last three changes in this browser', () => {
    const storage = memory();
    for (const id of ['a', 'b', 'c', 'd', 'b']) rememberChange(id, storage);
    expect(loadRecent(storage)).toEqual(['b', 'd', 'c']);
  });

  it('copes with broken or blocked storage', () => {
    expect(loadRecent({ getItem: () => '{oops' })).toEqual([]);
    expect(loadRecent({ getItem: () => '[1,"a"]' })).toEqual(['a']);
    const blocked = {
      getItem: () => {
        throw new Error('SecurityError');
      },
      setItem: () => {
        throw new Error('SecurityError');
      },
    };
    expect(rememberChange('a', blocked)).toEqual(['a']);
  });
});
