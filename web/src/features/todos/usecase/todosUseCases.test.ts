import { describe, expect, it, vi } from 'vitest';
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
    await uc.toggle({ id: '1', title: 't', due: '', done: false, overdue: false });
    expect(gateway.update).toHaveBeenCalledWith('1', 't', '', true);
  });
});
