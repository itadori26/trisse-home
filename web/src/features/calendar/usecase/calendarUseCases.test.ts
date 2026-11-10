import { describe, expect, it, vi } from 'vitest';
import { CalendarUseCases, type CalendarGateway } from './calendarUseCases';

describe('CalendarUseCases', () => {
  it('splits calendar urls by line', async () => {
    const gateway = { settings: vi.fn(), saveSettings: vi.fn(), events: vi.fn() } satisfies CalendarGateway;
    await new CalendarUseCases(gateway).saveSettings('https://a/\n\n https://b/ ', 'me', 'pw');
    expect(gateway.saveSettings).toHaveBeenCalledWith(['https://a/', 'https://b/'], 'me', 'pw');
  });
});
