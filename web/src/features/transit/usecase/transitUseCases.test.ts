import { describe, expect, it, vi } from 'vitest';
import { TransitUseCases, type TransitGateway } from './transitUseCases';

const gatewayMock = () =>
  ({
    stops: vi.fn().mockResolvedValue([]),
    stopsNear: vi.fn().mockResolvedValue([]),
    departures: vi.fn(),
  }) satisfies TransitGateway;

describe('TransitUseCases', () => {
  it('does not search for single characters', async () => {
    const gateway = gatewayMock();
    const uc = new TransitUseCases(gateway);
    await uc.searchStops(' a ');
    expect(gateway.stops).not.toHaveBeenCalled();
    await uc.searchStops(' Hbf ');
    expect(gateway.stops).toHaveBeenCalledWith('Hbf');
  });

  it('asks for the stops near a place by its coordinates (#694)', async () => {
    const gateway = gatewayMock();
    await new TransitUseCases(gateway).stopsNear({ lat: 53.144, lon: 8.215 });
    expect(gateway.stopsNear).toHaveBeenCalledWith(53.144, 8.215);
  });
});
