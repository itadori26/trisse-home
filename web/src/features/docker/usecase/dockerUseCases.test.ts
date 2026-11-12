import { describe, expect, it, vi } from 'vitest';
import { DockerUseCases, type DockerGateway } from './dockerUseCases';

describe('DockerUseCases', () => {
  it('loads the action capabilities once and retries after errors', async () => {
    const gateway = {
      containers: vi.fn(),
      actions: vi
        .fn()
        .mockRejectedValueOnce(new Error('offline'))
        .mockResolvedValue({ enabled: true, containers: ['homarr'] }),
      restart: vi.fn(),
      logs: vi.fn(),
      updates: vi.fn(),
      updateSettings: vi.fn(),
      saveUpdateSettings: vi.fn(),
    } satisfies DockerGateway;
    const uc = new DockerUseCases(gateway);
    await expect(uc.actions()).rejects.toThrow('offline');
    await uc.actions();
    await uc.actions();
    expect(gateway.actions).toHaveBeenCalledTimes(2);
  });
});
