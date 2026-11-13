import { describe, expect, it } from 'vitest';
import { artistLabel, coverUrl, parseWidgetConfig } from './releases';

describe('releases domain', () => {
  it('bounds the weeks', () => {
    expect(parseWidgetConfig({})).toEqual({ weeks: 4 });
    expect(parseWidgetConfig({ weeks: 40 })).toEqual({ weeks: 12 });
    expect(parseWidgetConfig({ weeks: 0 })).toEqual({ weeks: 1 });
  });
  it('labels artists and covers', () => {
    expect(artistLabel({ name: 'Low', disambiguation: 'US band' })).toBe('Low (US band)');
    expect(artistLabel({ name: 'Björk', disambiguation: '' })).toBe('Björk');
    expect(coverUrl('a b')).toBe('/api/releases/covers/a%20b');
  });
});
