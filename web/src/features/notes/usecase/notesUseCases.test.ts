import { describe, expect, it, vi } from 'vitest';
import { NotesUseCases, NoteValidationError, type NotesGateway } from './notesUseCases';

describe('NotesUseCases', () => {
  const gateway = {
    list: vi.fn().mockResolvedValue([]),
    save: vi.fn(),
    delete: vi.fn(),
    get: vi.fn(),
    search: vi.fn().mockResolvedValue([]),
    backlinks: vi.fn(),
    revisions: vi.fn(),
    revision: vi.fn(),
    restore: vi.fn(),
  } satisfies NotesGateway;
  const uc = new NotesUseCases(gateway);
  it('maps the scope to the pinned filter', async () => {
    await uc.list('pinned');
    expect(gateway.list).toHaveBeenCalledWith(true);
  });
  it('rejects empty notes', async () => {
    await expect(uc.save({ title: ' ', body: '', pinned: false })).rejects.toBeInstanceOf(
      NoteValidationError,
    );
    expect(gateway.save).not.toHaveBeenCalled();
  });
  it('searches only from two characters on (#73)', async () => {
    expect(await uc.search(' a ')).toEqual([]);
    expect(gateway.search).not.toHaveBeenCalled();
    await uc.search(' büro ');
    expect(gateway.search).toHaveBeenCalledWith('büro');
  });
});
