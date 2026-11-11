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
});
