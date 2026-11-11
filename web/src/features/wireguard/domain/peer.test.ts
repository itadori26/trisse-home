import { describe, expect, it } from 'vitest';
import { handshakeAgo } from './peer';

describe('handshakeAgo', () => {
  const now = new Date('2026-11-11T20:00:00Z');
  it.each([
    [null, 'never'],
    [new Date('2026-11-11T19:59:30Z'), 'just now'],
    [new Date('2026-11-11T19:55:00Z'), '5 min ago'],
    [new Date('2026-11-11T17:00:00Z'), '3 h ago'],
    [new Date('2026-11-08T20:00:00Z'), '3 d ago'],
  ])('%s → %s', (d, want) => expect(handshakeAgo(d, now)).toBe(want));
});
