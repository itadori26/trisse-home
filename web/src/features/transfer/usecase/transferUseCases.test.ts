import { describe, expect, it, vi } from 'vitest';
import { TransferUseCases, type TransferGateway } from './transferUseCases';

describe('TransferUseCases', () => {
  it('rejects foreign files before uploading', async () => {
    const gateway = { exportUrl: vi.fn(), import: vi.fn() } satisfies TransferGateway;
    await expect(new TransferUseCases(gateway).importFile('{"a":1}')).rejects.toThrow(/not a trisse-home/);
    expect(gateway.import).not.toHaveBeenCalled();
  });
});
