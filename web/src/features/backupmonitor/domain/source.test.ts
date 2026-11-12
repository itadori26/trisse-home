import { describe, expect, it } from 'vitest';
import { ago, looksLikeCron, pingExamples } from './source';

describe('backup monitor domain', () => {
  const now = new Date('2026-11-11T12:00:00Z');
  it.each([
    [null, 'never'],
    [new Date('2026-11-11T11:30:00Z'), '30 min ago'],
    [new Date('2026-11-11T02:00:00Z'), '10 h ago'],
    [new Date('2026-11-08T12:00:00Z'), '3 d ago'],
  ])('ago(%s) = %s', (d, want) => expect(ago(d, now)).toBe(want));
  it('builds curl examples', () => {
    const ex = pingExamples('http://h/api/checks/ping/x');
    expect(ex).toContain('/ping/x/start');
    expect(ex).toContain('/ping/x/fail');
  });

  it('recognises cron expressions', () => {
    expect(looksLikeCron('0 3 * * *')).toBe(true);
    expect(looksLikeCron('*/15 * * * 1-5')).toBe(true);
    expect(looksLikeCron('@daily')).toBe(true);
    expect(looksLikeCron('every day')).toBe(false);
  });
});
