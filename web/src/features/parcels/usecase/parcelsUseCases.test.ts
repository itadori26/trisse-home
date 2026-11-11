import { describe, expect, it, vi } from 'vitest';
import { ParcelsUseCases, type ParcelsGateway } from './parcelsUseCases';

describe('ParcelsUseCases', () => {
  it('validates before adding', async () => {
    const gateway = {
      settings: vi.fn(),
      saveSettings: vi.fn(),
      list: vi.fn(),
      add: vi.fn(),
      remove: vi.fn(),
    } satisfies ParcelsGateway;
    const uc = new ParcelsUseCases(gateway);
    await expect(uc.add('x', '')).rejects.toThrow(/8 to 35/);
    await uc.add('0034 0434 1234', ' Shoes ');
    expect(gateway.add).toHaveBeenCalledWith('003404341234', 'dhl', 'Shoes');
  });
});
