import { describe, expect, it } from 'vitest';
import { ago, pingExamples } from './source';

describe('backup monitor domain', () => {
  const now = new Date('2026-11-11T12:00:00Z');
  it.each([
    [null, 'never'],
    [new Date('2026-11-11T11:30:00Z'), '30 min ago'],
    [new Date('2026-11-11T02:00:00Z'), '10 h ago'],
    [new Date('2026-11-08T12:00:00Z'), '3 d ago'],
  ])('ago(%s) = %s', (d, want) => expect(ago(d, now)).toBe(want));
  it('builds curl examples', () => {
    expect(pingExamples('http://h/api/backup-status/ping/x')).toContain('?status=fail');
  });
});
