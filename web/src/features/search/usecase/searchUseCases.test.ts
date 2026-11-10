import { describe, expect, it, vi } from 'vitest';
import type { SearchGateway } from './ports';
import { SearchUseCases } from './searchUseCases';

function setup() {
  const gateway = {
    settings: vi.fn(),
    setDefaultProvider: vi.fn(),
    addShortcut: vi.fn(),
    deleteShortcut: vi.fn(),
    searchUrl: (q: string) => `/api/search?q=${encodeURIComponent(q)}`,
  } satisfies SearchGateway;
  return { uc: new SearchUseCases(gateway), gateway };
}

describe('SearchUseCases', () => {
  it('ignores blank queries', () => {
    expect(setup().uc.targetFor('   ')).toBeNull();
  });
  it('builds the backend search url', () => {
    expect(setup().uc.targetFor(' !w Oldenburg ')).toBe('/api/search?q=!w%20Oldenburg');
  });
  it('normalises shortcut keywords', async () => {
    const { uc, gateway } = setup();
    await uc.addShortcut({
      keyword: ' !GH ',
      name: 'GitHub',
      urlTemplate: 'https://github.com/search?q={query}',
    });
    expect(gateway.addShortcut).toHaveBeenCalledWith(expect.objectContaining({ keyword: 'gh' }));
  });
});
