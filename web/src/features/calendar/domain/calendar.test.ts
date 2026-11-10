import { describe, expect, it } from 'vitest';
import { type CalendarEvent, groupByDay, parseCalendarConfig, parseLocalDate } from './calendar';

const ev = (summary: string, start: Date, end: Date, allDay = false): CalendarEvent => ({
  uid: summary,
  summary,
  location: '',
  calendar: 'personal',
  start,
  end,
  allDay,
});

describe('groupByDay', () => {
  const today = new Date(2026, 10, 10, 9, 0);
  it('groups by local day and spans multi-day events', () => {
    const events = [
      ev('meeting', new Date(2026, 10, 10, 14), new Date(2026, 10, 10, 15)),
      ev('trip', parseLocalDate('2026-11-11'), parseLocalDate('2026-11-13'), true),
    ];
    const days = groupByDay(events, today, 7);
    expect(days.map((d) => [d.key, d.events.map((e) => e.summary)])).toEqual([
      ['2026-11-10', ['meeting']],
      ['2026-11-11', ['trip']],
      ['2026-11-12', ['trip']],
    ]);
  });
  it('respects the number of days', () => {
    const events = [ev('later', new Date(2026, 10, 20, 10), new Date(2026, 10, 20, 11))];
    expect(groupByDay(events, today, 7)).toEqual([]);
  });
  it('clamps the configured days', () => {
    expect(parseCalendarConfig({ days: 100 }).days).toBe(31);
    expect(parseCalendarConfig({}).days).toBe(7);
  });
});
