import { describe, expect, it, vi } from 'vitest';
import { BookmarksUseCases, type BookmarksGateway } from './bookmarksUseCases';

describe('BookmarksUseCases', () => {
  const gateway = {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    createFolder: vi.fn(),
    renameFolder: vi.fn(),
    deleteFolder: vi.fn(),
    importHtml: vi.fn(),
    getSettings: vi.fn(),
    saveSettings: vi.fn(),
    enrich: vi.fn(),
    checkLink: vi.fn(),
  } satisfies BookmarksGateway;
  const uc = new BookmarksUseCases(gateway);
  const input = { folderId: '', title: '', url: 'go.dev', description: '', tags: [] };

  it('adds https:// to bare hosts', async () => {
    await uc.save(input);
    expect(gateway.create).toHaveBeenCalledWith(expect.objectContaining({ url: 'https://go.dev' }));
  });
  it('keeps explicit schemes and updates with an id', async () => {
    await uc.save({ ...input, url: 'http://nas.lan' }, 'b1');
    expect(gateway.update).toHaveBeenCalledWith('b1', expect.objectContaining({ url: 'http://nas.lan' }));
  });
  it('passes enrich, check and settings to the backend (#72)', async () => {
    await uc.enrich('b1');
    await uc.checkLink('b1');
    await uc.saveSettings({ enrich: true, checkLinks: false });
    expect(gateway.enrich).toHaveBeenCalledWith('b1');
    expect(gateway.checkLink).toHaveBeenCalledWith('b1');
    expect(gateway.saveSettings).toHaveBeenCalledWith({ enrich: true, checkLinks: false });
  });
});
