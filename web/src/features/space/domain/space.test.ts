import { describe, expect, it } from 'vitest';
import { backgroundUrl, countdown, parseSpaceConfig, statusTone } from './space';

describe('space domain (#64)', () => {
  it('reads the widget config', () => {
    expect(parseSpaceConfig({})).toEqual({ show: 'both', count: 4 });
    expect(parseSpaceConfig({ show: 'picture', count: 99 })).toEqual({ show: 'picture', count: 10 });
    expect(parseSpaceConfig({ show: 'x', count: 0 })).toEqual({ show: 'both', count: 1 });
  });

  it('counts down to a launch', () => {
    const now = new Date(2026, 10, 13, 12, 0);
    const at = (minutes: number) => new Date(now.getTime() + minutes * 60_000);
    expect(countdown(at(-5), now)).toBe('now');
    expect(countdown(at(25), now)).toBe('in 25 min');
    expect(countdown(at(5 * 60), now)).toBe('in 5 h');
    expect(countdown(at(3 * 24 * 60), now)).toBe('in 3 days');
  });

  it('colours statuses and builds the background address', () => {
    expect(statusTone('go')).toBe('good');
    expect(statusTone('hold')).toBe('bad');
    expect(statusTone('tbd')).toBe('neutral');
    expect(backgroundUrl('http://home:8090')).toBe('http://home:8090/api/space/apod/image');
  });
});
