import { describe, expect, it } from 'vitest';
import { type BackupStatus, dayWord, downloadUrl, healthLine } from './backup';

describe('downloadUrl', () => {
  it('encodes the file name', () => {
    expect(downloadUrl({ name: 'waylight-20261111-030000.db', size: 1, createdAt: new Date() })).toBe(
      '/api/backups/waylight-20261111-030000.db',
    );
  });
});

describe('healthLine (#677)', () => {
  const base: BackupStatus = {
    health: 'ok',
    lastBackupAt: new Date(2026, 10, 21, 3, 0),
    count: 3,
    failedAt: null,
    error: '',
    offsiteFailedAt: null,
    offsiteError: '',
  };

  it('puts a failure before everything else', () => {
    const failedAt = new Date(2026, 10, 21, 4, 0);
    expect(healthLine({ ...base, health: 'failed', failedAt, offsiteFailedAt: failedAt })).toEqual({
      tone: 'error',
      key: 'backup.health.failed',
      at: failedAt,
    });
  });

  it('tells a missing backup, a failed off-site copy and an old backup apart', () => {
    expect(healthLine({ ...base, health: 'stale', lastBackupAt: null }).key).toBe('backup.health.none');
    expect(healthLine({ ...base, health: 'stale', offsiteFailedAt: new Date() }).key).toBe(
      'backup.health.offsite_failed',
    );
    expect(healthLine({ ...base, health: 'stale' })).toMatchObject({
      tone: 'warn',
      key: 'backup.health.stale',
    });
    expect(healthLine(base)).toEqual({ tone: 'ok', key: 'backup.health.ok', at: base.lastBackupAt });
  });
});

describe('dayWord', () => {
  const now = new Date(2026, 10, 21, 12, 0);
  it('names today and yesterday', () => {
    expect(dayWord(new Date(2026, 10, 21, 3, 0), now)).toBe('today');
    expect(dayWord(new Date(2026, 10, 20, 23, 0), now)).toBe('yesterday');
    expect(dayWord(new Date(2026, 10, 19, 23, 0), now)).toBe('earlier');
  });
});
