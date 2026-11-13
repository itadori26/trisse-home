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
