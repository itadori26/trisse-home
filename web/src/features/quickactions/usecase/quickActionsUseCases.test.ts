import { describe, expect, it, vi } from 'vitest';
import { searchPrefixes } from '../domain/command';
import { QuickActionError, QuickActionsUseCases, type QuickTargets } from './quickActionsUseCases';

function setup() {
  const targets = {
    addTodo: vi.fn().mockResolvedValue({ id: 't1' }),
    deleteTodo: vi.fn().mockResolvedValue(undefined),
    addNote: vi.fn().mockResolvedValue({ id: 'n1' }),
    deleteNote: vi.fn().mockResolvedValue(undefined),
    addBookmark: vi.fn().mockResolvedValue({ id: 'b1' }),
    deleteBookmark: vi.fn().mockResolvedValue(undefined),
    restartableContainers: vi.fn().mockResolvedValue(['nextcloud']),
    restartContainer: vi.fn().mockResolvedValue(undefined),
    wakeableServices: vi.fn().mockResolvedValue([{ id: 's1', name: 'NAS' }]),
    wakeService: vi.fn().mockResolvedValue(undefined),
  } satisfies QuickTargets;
  const uc = new QuickActionsUseCases(targets, { now: () => new Date(2026, 10, 12, 9) });
  return { uc, targets };
}

describe('QuickActionsUseCases', () => {
  it('creates items and undoes them', async () => {
    const { uc, targets } = setup();
    const cmd = uc.parse('t Müll morgen', searchPrefixes);
    expect(cmd).toEqual({ kind: 'todo', title: 'Müll', due: '2026-11-13' });
    if (!cmd) throw new Error('parse');
    const res = await uc.execute(cmd);
    expect(targets.addTodo).toHaveBeenCalledWith('Müll', '2026-11-13');
    await res.undo?.();
    expect(targets.deleteTodo).toHaveBeenCalledWith('t1');

    const bm = await uc.execute({ kind: 'bookmark', url: 'https://x.org/', title: 'x.org' });
    expect(bm.openUrl).toBe('https://x.org/');
    expect(uc.today()).toBe('2026-11-12');
  });

  it('only runs allow-listed actions', async () => {
    const { uc, targets } = setup();
    expect((await uc.execute({ kind: 'restart', target: 'NEXTCLOUD' })).command).toEqual({
      kind: 'restart',
      target: 'nextcloud',
    });
    expect(targets.restartContainer).toHaveBeenCalledWith('nextcloud');
    await expect(uc.execute({ kind: 'restart', target: 'postgres' })).rejects.toBeInstanceOf(
      QuickActionError,
    );
    await uc.execute({ kind: 'wake', target: 'nas' });
    expect(targets.wakeService).toHaveBeenCalledWith('s1');
    await expect(uc.execute({ kind: 'wake', target: 'router' })).rejects.toBeInstanceOf(QuickActionError);
  });
});
