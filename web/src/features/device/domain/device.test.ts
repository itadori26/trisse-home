import { describe, expect, it } from 'vitest';
import { changing } from './device';

describe('device (#140)', () => {
  it('polls only while the system works on a change', () => {
    const s = { requested: 'kiosk', error: '', display: 'unknown' } as const;
    expect(changing({ ...s, phase: 'pending' })).toBe(true);
    expect(changing({ ...s, phase: 'installing' })).toBe(true);
    expect(changing({ ...s, phase: 'active' })).toBe(false);
    expect(changing({ ...s, phase: 'failed' })).toBe(false);
  });
});
