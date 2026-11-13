import { describe, expect, it, vi } from 'vitest';
import { emptyServiceInput, type Service } from '../domain/service';
import { ServicesUseCases, ServiceValidationError, type ServicesGateway } from './servicesUseCases';

const svc = (id: string): Service => ({
  ...emptyServiceInput(),
  id,
  name: id,
  url: 'https://x',
  position: 0,
  state: 'unknown',
  httpStatus: 0,
  latencyMs: 0,
  checkedAt: null,
  error: '',
  certExpiresAt: null,
  certExpiresSoon: false,
  container: '',
  hidden: false,
  mac: '',
});

function setup() {
  const gateway = {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    reorder: vi.fn().mockResolvedValue(undefined),
    checkNow: vi.fn(),
    history: vi.fn(),
    wake: vi.fn(),
    uptime: vi.fn(),
    incidents: vi.fn(),
    discovery: vi.fn(),
    saveDiscovery: vi.fn(),
    syncDiscovery: vi.fn(),
  } satisfies ServicesGateway;
  return { uc: new ServicesUseCases(gateway), gateway };
}

describe('ServicesUseCases', () => {
  it('validates before saving', async () => {
    const { uc, gateway } = setup();
    await expect(uc.save(emptyServiceInput())).rejects.toBeInstanceOf(ServiceValidationError);
    expect(gateway.create).not.toHaveBeenCalled();
  });
  it('updates when an id is given', async () => {
    const { uc, gateway } = setup();
    await uc.save({ ...emptyServiceInput(), name: 'a', url: 'https://a' }, '7');
    expect(gateway.update).toHaveBeenCalledWith('7', expect.anything());
  });
  it('moves services and ignores moves past the ends', async () => {
    const { uc, gateway } = setup();
    const list = [svc('a'), svc('b')];
    expect((await uc.move(list, 'b', -1)).map((s) => s.id)).toEqual(['b', 'a']);
    expect(gateway.reorder).toHaveBeenCalledWith(['b', 'a']);
    expect((await uc.move(list, 'a', -1)).map((s) => s.id)).toEqual(['a', 'b']);
    expect(gateway.reorder).toHaveBeenCalledOnce();
  });
  it('hides tiles of removed containers on the dashboard only', async () => {
    const { uc, gateway } = setup();
    gateway.list.mockResolvedValue([svc('a'), { ...svc('b'), container: 'b', hidden: true }]);
    gateway.checkNow.mockResolvedValue([{ ...svc('b'), hidden: true }]);
    expect((await uc.list()).map((s) => s.id)).toEqual(['a']);
    expect((await uc.listAll()).map((s) => s.id)).toEqual(['a', 'b']);
    expect(await uc.checkNow()).toEqual([]);
  });
  it('validates the discovery interval before saving', async () => {
    const { uc, gateway } = setup();
    await expect(
      uc.saveDiscovery({ enabled: true, intervalMinutes: 0, rule: 'labels' }),
    ).rejects.toBeInstanceOf(ServiceValidationError);
    await expect(
      uc.saveDiscovery({ enabled: true, intervalMinutes: 2.5, rule: 'labels' }),
    ).rejects.toBeInstanceOf(ServiceValidationError);
    expect(gateway.saveDiscovery).not.toHaveBeenCalled();
    await uc.saveDiscovery({ enabled: true, intervalMinutes: 15, rule: 'manual' });
    expect(gateway.saveDiscovery).toHaveBeenCalledWith({
      enabled: true,
      intervalMinutes: 15,
      rule: 'manual',
    });
  });
});
