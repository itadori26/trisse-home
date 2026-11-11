import { describe, expect, it, vi } from 'vitest';
import { ShoppingUseCases, type ShoppingGateway } from './shoppingUseCases';

describe('ShoppingUseCases', () => {
  it('ignores empty input and trims items', async () => {
    const gateway = {
      settings: vi.fn(),
      saveSettings: vi.fn(),
      items: vi.fn(),
      add: vi.fn().mockResolvedValue({ id: 'a', summary: 'Milk', done: false }),
      setDone: vi.fn(),
      remove: vi.fn(),
      clearCompleted: vi.fn(),
    } satisfies ShoppingGateway;
    const uc = new ShoppingUseCases(gateway);
    expect(await uc.add('   ')).toBeNull();
    await uc.add(' Milk ');
    expect(gateway.add).toHaveBeenCalledWith('Milk');
  });
});
