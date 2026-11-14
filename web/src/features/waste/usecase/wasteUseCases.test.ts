import { describe, expect, it, vi } from 'vitest';
import { WasteUseCases, type WasteGateway } from './wasteUseCases';

const gw = () =>
  ({
    settings: vi.fn(),
    saveSettings: vi.fn(),
    uploadCalendar: vi.fn(),
    removeCalendar: vi.fn(),
    upcoming: vi.fn(),
  }) satisfies WasteGateway;

describe('WasteUseCases', () => {
  it('drops empty rules before saving', async () => {
    const gateway = gw();
    await new WasteUseCases(gateway).saveSettings(' https://x/a.ics ', [
      { match: ' bio ', label: '', color: '', icon: '' },
      { match: ' ', label: 'x', color: '', icon: '' },
    ]);
    expect(gateway.saveSettings).toHaveBeenCalledWith('https://x/a.ics', [
      { match: 'bio', label: '', color: '', icon: '' },
    ]);
  });

  it('checks a calendar file before uploading it (#153)', async () => {
    const gateway = gw();
    const uc = new WasteUseCases(gateway);
    const file = (name: string, size: number, type = '') => new File([new Uint8Array(size)], name, { type });
    await expect(uc.uploadCalendar(file('plan.pdf', 10, 'application/pdf'))).rejects.toThrow();
    await expect(uc.uploadCalendar(file('big.ics', (1 << 20) + 1))).rejects.toThrow();
    await expect(uc.uploadCalendar(file('empty.ics', 0))).rejects.toThrow();
    expect(gateway.uploadCalendar).not.toHaveBeenCalled();
    const ok = file('abfall-2027.ics', 100);
    await uc.uploadCalendar(ok);
    await uc.uploadCalendar(file('download', 100, 'text/calendar;charset=utf-8'));
    expect(gateway.uploadCalendar).toHaveBeenNthCalledWith(1, ok);
    expect(gateway.uploadCalendar).toHaveBeenCalledTimes(2);
  });

  it('keeps the source when saving rules without a link', async () => {
    const gateway = gw();
    await new WasteUseCases(gateway).saveSettings('  ', [{ match: 'bio', label: '', color: '', icon: '' }]);
    expect(gateway.saveSettings).toHaveBeenCalledWith('', [{ match: 'bio', label: '', color: '', icon: '' }]);
  });
});
