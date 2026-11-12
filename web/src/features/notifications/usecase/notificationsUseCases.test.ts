import { describe, expect, it, vi } from 'vitest';
import { NotificationsUseCases, type NotificationsGateway } from './notificationsUseCases';

describe('NotificationsUseCases', () => {
  it('saves before sending a test', async () => {
    const calls: string[] = [];
    const gateway = {
      settings: vi.fn(),
      save: vi.fn(async () => {
        calls.push('save');
        return {} as never;
      }),
      sendTest: vi.fn(async () => {
        calls.push('test');
      }),
    } satisfies NotificationsGateway;
    const rules = {
      services: true,
      containers: true,
      disk: true,
      diskThreshold: 90,
      tls: true,
      tlsDays: 14,
      backups: true,
      security: true,
    };
    await new NotificationsUseCases(gateway).saveAndTest({
      enabled: true,
      serverUrl: ' https://ntfy.sh ',
      topic: ' t ',
      token: '',
      rules,
    });
    expect(calls).toEqual(['save', 'test']);
    expect(gateway.save).toHaveBeenCalledWith(
      expect.objectContaining({ serverUrl: 'https://ntfy.sh', topic: 't' }),
    );
  });
});
