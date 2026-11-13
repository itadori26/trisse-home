import { describe, expect, it, vi } from 'vitest';
import { SmarthomeUseCases, type SmartHomeGateway } from './smarthomeUseCases';

const gateway = () =>
  ({
    settings: vi.fn(),
    saveSettings: vi.fn(),
    entities: vi.fn(),
    states: vi.fn().mockResolvedValue([]),
    act: vi.fn(),
  }) satisfies SmartHomeGateway;

describe('SmarthomeUseCases', () => {
  it('asks nothing for a widget without entities', async () => {
    const g = gateway();
    expect(await new SmarthomeUseCases(g).states([])).toEqual([]);
    expect(g.states).not.toHaveBeenCalled();
  });
  it('trims url and token', async () => {
    const g = gateway();
    await new SmarthomeUseCases(g).saveSettings(' http://ha:8123 ', ' tok ');
    expect(g.saveSettings).toHaveBeenCalledWith('http://ha:8123', 'tok');
  });
});
