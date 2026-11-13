import { describe, expect, it, vi } from 'vitest';
import { ArrUseCases, type ArrGateway } from './arrUseCases';

const gateway = () =>
  ({
    instances: vi.fn(),
    add: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    overview: vi.fn(),
  }) satisfies ArrGateway;

describe('ArrUseCases', () => {
  it('validates and trims before saving', async () => {
    const g = gateway();
    const uc = new ArrUseCases(g);
    await expect(uc.save({ app: 'sonarr', name: '', url: 'x', apiKey: 'k' })).rejects.toThrow();
    await uc.save({ app: 'radarr', name: ' Films ', url: ' http://r ', apiKey: ' k ' });
    expect(g.add).toHaveBeenCalledWith({ app: 'radarr', name: 'Films', url: 'http://r', apiKey: 'k' });
    await uc.save({ app: 'radarr', name: '', url: 'http://r', apiKey: '' }, 'id1');
    expect(g.update).toHaveBeenCalledWith('id1', { app: 'radarr', name: '', url: 'http://r', apiKey: '' });
  });
});
