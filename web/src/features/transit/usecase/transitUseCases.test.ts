import { describe, expect, it, vi } from 'vitest';
import { TransitUseCases, type TransitGateway } from './transitUseCases';

describe('TransitUseCases', () => {
  it('does not search for single characters', async () => {
    const gateway = { stops: vi.fn().mockResolvedValue([]), departures: vi.fn() } satisfies TransitGateway;
    const uc = new TransitUseCases(gateway);
    await uc.searchStops(' a ');
    expect(gateway.stops).not.toHaveBeenCalled();
    await uc.searchStops(' Hbf ');
    expect(gateway.stops).toHaveBeenCalledWith('Hbf');
  });
});
