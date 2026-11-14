import { describe, expect, it, vi } from 'vitest';
import { buildAiContext } from './aiContext';
import type { Container } from './container';

describe('AI context (#162)', () => {
  it('lists only status lines of enabled modules and skips failing sources', async () => {
    const c = {
      system: {
        snapshot: vi.fn().mockResolvedValue({ cpuPercent: 12.4, memory: { percent: 40.6 }, load: [0.5] }),
      },
      services: {
        list: vi.fn().mockResolvedValue([
          { name: 'Jellyfin', state: 'offline', error: 'timeout' },
          { name: 'Nextcloud', state: 'online', error: '' },
        ]),
      },
      docker: { containers: vi.fn().mockRejectedValue(new Error('no socket')) },
      timeline: { list: vi.fn() },
    } as unknown as Container;
    const ctx = await buildAiContext(
      c,
      new Set(['system', 'services', 'docker']),
      new Date('2026-11-14T08:00:00Z'),
    );
    expect(ctx).toBe(
      'now: 2026-11-14 08:00 UTC\nhost: cpu 12 %, memory 41 %, load 0.50\nservice Jellyfin: offline (timeout)\nservice Nextcloud: online',
    );
    expect(c.timeline.list).not.toHaveBeenCalled();
  });
});
