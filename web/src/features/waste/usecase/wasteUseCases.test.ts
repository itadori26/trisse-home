import { describe, expect, it, vi } from 'vitest';
import { WasteUseCases, type WasteGateway } from './wasteUseCases';

describe('WasteUseCases', () => {
  it('drops empty rules before saving', async () => {
    const gateway = { settings: vi.fn(), saveSettings: vi.fn(), upcoming: vi.fn() } satisfies WasteGateway;
    await new WasteUseCases(gateway).saveSettings(' https://x/a.ics ', [
      { match: ' bio ', label: '', color: '', icon: '' },
      { match: ' ', label: 'x', color: '', icon: '' },
    ]);
    expect(gateway.saveSettings).toHaveBeenCalledWith('https://x/a.ics', [
      { match: 'bio', label: '', color: '', icon: '' },
    ]);
  });
});
