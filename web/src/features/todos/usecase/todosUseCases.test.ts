import { describe, expect, it, vi } from 'vitest';
import { NO_REMINDER, NO_REPEAT } from '../domain/todo';
import { TodosUseCases, type TodosGateway } from './todosUseCases';

describe('TodosUseCases', () => {
  const gateway = {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    clearCompleted: vi.fn(),
  } satisfies TodosGateway;
  const uc = new TodosUseCases(gateway);

  it('extracts a due date suffix', async () => {
    await uc.add('Pay rent @2026-12-01');
    expect(gateway.create).toHaveBeenCalledWith('Pay rent', '2026-12-01');
  });
  it('keeps text without due date', async () => {
    await uc.add(' Email @Alex ');
    expect(gateway.create).toHaveBeenCalledWith('Email @Alex', '');
  });
  it('toggles completion', async () => {
    await uc.toggle({
      id: '1',
      title: 't',
      due: '',
      done: false,
      overdue: false,
      version: 1,
      repeat: NO_REPEAT,
      reminder: NO_REMINDER,
    });
    expect(gateway.update).toHaveBeenCalledWith('1', 't', '', true);
  });
  it('saves recurrence and reminder after validation (#68)', async () => {
    const todo = {
      id: '1',
      title: 't',
      due: '2026-11-12',
      done: false,
      overdue: false,
      version: 1,
      repeat: NO_REPEAT,
      reminder: NO_REMINDER,
    };
    await uc.editFull(todo, {
      title: ' Bins ',
      due: '2026-11-12',
      repeat: { kind: 'weekly', every: 1, days: [4, 1, 4] },
      reminder: { at: '19:00', daysBefore: 0 },
    });
    expect(gateway.update).toHaveBeenLastCalledWith('1', 'Bins', '2026-11-12', false, {
      repeat: { kind: 'weekly', every: 1, days: [1, 4] },
      reminder: { at: '19:00', daysBefore: 0 },
    });
    await expect(
      uc.editFull(todo, { title: 't', due: '', repeat: NO_REPEAT, reminder: { at: '08:00', daysBefore: 0 } }),
    ).rejects.toThrow(/due date/);
  });
});
