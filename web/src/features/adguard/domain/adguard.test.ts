import { describe, expect, it } from 'vitest';
import { needsUsername, parseWidgetConfig, providerName } from './adguard';

describe('DNS filter domain', () => {
  it('keeps widgets from before Pi-hole support on AdGuard Home', () => {
    expect(parseWidgetConfig({})).toEqual({ provider: 'adguard' });
    expect(parseWidgetConfig({ provider: 'bind' })).toEqual({ provider: 'adguard' });
    expect(parseWidgetConfig({ provider: 'pihole' })).toEqual({ provider: 'pihole' });
  });
  it('names providers and their login fields', () => {
    expect(providerName('pihole')).toBe('Pi-hole');
    expect(needsUsername('pihole')).toBe(false);
    expect(needsUsername('adguard')).toBe(true);
  });
});

describe('pause countdown (#177)', () => {
  it('rounds the minutes left up and stops at zero', async () => {
    const { pauseMinutesLeft } = await import('./adguard');
    const now = new Date(2026, 10, 14, 8, 0, 0);
    expect(pauseMinutesLeft(null, now)).toBe(0);
    expect(pauseMinutesLeft(new Date(now.getTime() + 29.5 * 60_000), now)).toBe(30);
    expect(pauseMinutesLeft(new Date(now.getTime() + 10_000), now)).toBe(1);
    expect(pauseMinutesLeft(new Date(now.getTime() - 1), now)).toBe(0);
  });
});
