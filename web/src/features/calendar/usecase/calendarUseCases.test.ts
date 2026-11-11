import { describe, expect, it, vi } from 'vitest';
import { CalendarUseCases, type CalendarGateway } from './calendarUseCases';

const gateway = () =>
  ({
    settings: vi.fn(),
    saveSettings: vi.fn(),
    events: vi.fn(),
    range: vi.fn(),
    create: vi.fn(),
  }) satisfies CalendarGateway;

describe('CalendarUseCases', () => {
  it('splits calendar URLs', async () => {
    const g = gateway();
    await new CalendarUseCases(g).saveSettings(' https://a/ \n https://b/ ', ' me ', 'pw');
    expect(g.saveSettings).toHaveBeenCalledWith(['https://a/', 'https://b/'], 'me', 'pw');
  });

  it('sends timed events as instants and all-day events as dates', async () => {
    const g = gateway();
    const uc = new CalendarUseCases(g);
    const base = {
      calendarUrl: 'u',
      summary: ' Dentist ',
      location: '',
      date: '2026-11-12',
      endDate: '',
      startTime: '09:00',
      endTime: '10:00',
    };
    await uc.createEvent({ ...base, allDay: false });
    expect(g.create).toHaveBeenLastCalledWith(
      expect.objectContaining({ summary: 'Dentist', start: new Date('2026-11-12T09:00').toISOString() }),
    );
    await uc.createEvent({ ...base, allDay: true });
    expect(g.create).toHaveBeenLastCalledWith(
      expect.objectContaining({ start: '2026-11-12', end: '2026-11-12' }),
    );
    await expect(uc.createEvent({ ...base, allDay: false, summary: '' })).rejects.toThrow(/title/);
  });
});
