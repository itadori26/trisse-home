import { describe, expect, it } from 'vitest';
import { activeMaintenance, suggestTopic } from './notifications';

describe('suggestTopic', () => {
  it('creates a long, valid topic', () => {
    const t = suggestTopic(() => 0.5);
    expect(t).toMatch(/^trisse-home-[a-z0-9]{12}$/);
  });
});

describe('maintenance', () => {
  it('finds active windows', () => {
    const now = new Date('2026-11-12T10:00:00Z');
    const w = (start: string, end: string) => ({
      id: start,
      scope: ['*'],
      start: new Date(start),
      end: new Date(end),
      note: '',
    });
    const list = [
      w('2026-11-12T09:00:00Z', '2026-11-12T11:00:00Z'),
      w('2026-11-12T12:00:00Z', '2026-11-12T13:00:00Z'),
    ];
    expect(activeMaintenance(list, now).map((x) => x.id)).toEqual(['2026-11-12T09:00:00Z']);
  });
});
