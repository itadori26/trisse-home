import { describe, expect, it, vi } from 'vitest';
import { NotesUseCases, NoteValidationError, type NotesGateway } from './notesUseCases';

describe('NotesUseCases', () => {
  const gateway = {
    list: vi.fn().mockResolvedValue([]),
    save: vi.fn(),
    delete: vi.fn(),
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
});
