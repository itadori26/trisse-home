import { describe, expect, it } from 'vitest';
import { downloadUrl } from './backup';

describe('downloadUrl', () => {
  it('encodes the file name', () => {
    expect(downloadUrl({ name: 'trisse-home-20261111-030000.db', size: 1, createdAt: new Date() })).toBe(
      '/api/backups/trisse-home-20261111-030000.db',
    );
  });
});
