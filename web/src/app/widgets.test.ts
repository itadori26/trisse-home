import { describe, expect, it } from 'vitest';
import { widgetDefinitions } from './widgets';

describe('widget registry (#116)', () => {
  it('loads every view and settings component lazily and caches the chunk', async () => {
    for (const def of widgetDefinitions) {
      for (const r of [def.view, def.settings]) {
        if (!r) continue;
        const first = await r.load();
        expect(first, def.type.kind).toBeTypeOf('function');
        expect(await r.load()).toBe(first);
      }
    }
  });

  it('has unique kinds', () => {
    const kinds = widgetDefinitions.map((d) => d.type.kind);
    expect(new Set(kinds).size).toBe(kinds.length);
  });
});
