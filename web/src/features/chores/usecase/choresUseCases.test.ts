import { describe, expect, it, vi } from 'vitest';
import { ChoresUseCases, type ChoresGateway } from './choresUseCases';

describe('ChoresUseCases', () => {
  it('sends assignees as a list and ticks chores off', async () => {
    const g: ChoresGateway = {
      list: vi.fn(async () => []),
      create: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(async () => {}),
      done: vi.fn(),
      shareGroups: vi.fn(async () => []),
    };
    const uc = new ChoresUseCases(g);
    await uc.save(null, { title: ' Dishes ', everyDays: 1, due: '2026-11-23', assignees: 'Anna\nBen' });
    expect(g.create).toHaveBeenCalledWith(
      { title: 'Dishes', everyDays: 1, due: '2026-11-23', assignees: ['Anna', 'Ben'] },
      undefined,
    );
    await uc.save('1', { title: 'Dishes', everyDays: 2, due: '2026-11-23', assignees: '' }, 'private');
    expect(g.update).toHaveBeenCalledWith(
      '1',
      expect.objectContaining({ everyDays: 2, assignees: [] }),
      'private',
    );
    await uc.done('1');
    expect(g.done).toHaveBeenCalledWith('1');
  });
});
