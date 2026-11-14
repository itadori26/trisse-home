import { describe, expect, it, vi } from 'vitest';
import { FritzBoxUseCases, type FritzBoxGateway } from './fritzboxUseCases';

describe('FritzBoxUseCases', () => {
  it('trims URL and user but not the password', async () => {
    const gateway = {
      settings: vi.fn(),
      saveSettings: vi.fn(),
      status: vi.fn(),
      calls: vi.fn(),
    } satisfies FritzBoxGateway;
    await new FritzBoxUseCases(gateway).saveSettings(' http://fritz.box ', ' home ', ' pw ');
    expect(gateway.saveSettings).toHaveBeenCalledWith('http://fritz.box', 'home', ' pw ');
  });
  it('asks for the calls with the widget filter (#168)', async () => {
    const gateway = {
      settings: vi.fn(),
      saveSettings: vi.fn(),
      status: vi.fn(),
      calls: vi.fn(),
    } satisfies FritzBoxGateway;
    await new FritzBoxUseCases(gateway).calls({ filter: 'missed', limit: 20, showNumbers: false });
    expect(gateway.calls).toHaveBeenCalledWith('missed', 20);
  });
});
