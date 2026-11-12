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
      channels: vi.fn(),
      saveChannel: vi.fn(),
      deleteChannel: vi.fn(),
      testChannel: vi.fn(),
      maintenance: vi.fn(),
      startMaintenance: vi.fn(),
      endMaintenance: vi.fn(),
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
      policy: { severities: {}, quiet: { enabled: false, start: '22:00', end: '07:00', days: [] } },
    });
    expect(calls).toEqual(['save', 'test']);
    expect(gateway.save).toHaveBeenCalledWith(
      expect.objectContaining({ serverUrl: 'https://ntfy.sh', topic: 't' }),
    );
  });

  it('parses the maintenance scope and starts at least one minute', async () => {
    const startMaintenance = vi.fn();
    const uc = new NotificationsUseCases({ startMaintenance } as unknown as NotificationsGateway);
    await uc.startMaintenance(' service:NAS , ,container:nginx ', 0.2, ' update ');
    expect(startMaintenance).toHaveBeenCalledWith(['service:NAS', 'container:nginx'], 1, 'update');
  });
});
