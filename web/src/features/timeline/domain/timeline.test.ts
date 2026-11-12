import { describe, expect, it } from 'vitest';
import { groupByDay, timelineConfig, type TimelineEvent } from './timeline';

const ev = (id: string, at: string): TimelineEvent => ({
  id,
  at: new Date(at),
  source: 'service',
  severity: 'info',
  subject: '',
  message: id,
});

describe('timeline', () => {
  it('groups by day keeping the order', () => {
    const groups = groupByDay(
      [ev('c', '2026-11-12T10:00:00Z'), ev('b', '2026-11-12T08:00:00Z'), ev('a', '2026-11-11T23:00:00Z')],
      (d) => d.toISOString().slice(0, 10),
    );
    expect(groups.map((g) => [g.day, g.events.map((e) => e.id)])).toEqual([
      ['2026-11-12', ['c', 'b']],
      ['2026-11-11', ['a']],
    ]);
  });

  it('reads the widget config with bounds', () => {
    expect(timelineConfig({})).toEqual({ compact: true, limit: 10 });
    expect(timelineConfig({ compact: false, limit: 500 })).toEqual({ compact: false, limit: 100 });
  });
});
