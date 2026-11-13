import { describe, expect, it } from 'vitest';
import { appName, defaultPort, emptyInstance, parseWidgetConfig, validateInstance } from './arr';

describe('arr domain', () => {
  it('shows every section by default', () => {
    expect(parseWidgetConfig({})).toEqual({ upcoming: true, queue: true, health: true });
    expect(parseWidgetConfig({ queue: false }).queue).toBe(false);
  });
  it('requires a URL and, for new instances, an API key', () => {
    const base = { ...emptyInstance(), url: 'http://sonarr:8989', apiKey: 'k' };
    expect(validateInstance(base, true)).toBeNull();
    expect(validateInstance({ ...base, apiKey: '' }, true)).not.toBeNull();
    expect(validateInstance({ ...base, apiKey: '' }, false)).toBeNull();
    expect(validateInstance({ ...base, url: 'sonarr' }, false)).not.toBeNull();
  });
  it('names apps and their default ports', () => {
    expect(appName('prowlarr')).toBe('Prowlarr');
    expect(defaultPort('radarr')).toBe(7878);
  });
});
