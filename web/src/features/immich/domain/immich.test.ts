import { describe, expect, it } from 'vitest';
import { parseWidgetConfig, thumbnailUrl } from './immich';

describe('immich domain', () => {
  it('builds proxied thumbnail URLs and shows memories by default', () => {
    expect(thumbnailUrl('a/b')).toBe('/api/immich/thumbnails/a%2Fb');
    expect(parseWidgetConfig({})).toEqual({ showMemories: true });
    expect(parseWidgetConfig({ showMemories: false })).toEqual({ showMemories: false });
  });
});
