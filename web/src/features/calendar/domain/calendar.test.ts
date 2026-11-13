import { describe, expect, it } from 'vitest';
import {
  type CalendarEvent,
  calendarName,
  eventsOnDay,
  groupByDay,
  monthGrid,
  parseCalendarConfig,
  parseLocalDate,
  validateNewEvent,
  validateSubscription,
} from './calendar';

const ev = (summary: string, start: Date, end: Date, allDay = false): CalendarEvent => ({
  uid: summary,
  summary,
  location: '',
  calendar: 'personal',
  color: '',
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

  describe('monthGrid', () => {
    it('starts on Monday and covers six weeks', () => {
      const cells = monthGrid(2026, 10, new Date(2026, 10, 11)); // November 2026 starts on a Sunday
      expect(cells).toHaveLength(42);
      expect(cells[0]?.key).toBe('2026-10-26');
      expect(cells[6]?.key).toBe('2026-11-01');
      expect(cells.filter((c) => c.inMonth)).toHaveLength(30);
      expect(cells.find((c) => c.isToday)?.key).toBe('2026-11-11');
    });
    it('handles leap years', () => {
      expect(monthGrid(2028, 1, new Date()).filter((c) => c.inMonth)).toHaveLength(29);
      expect(monthGrid(2026, 1, new Date()).filter((c) => c.inMonth)).toHaveLength(28);
    });
  });

  it('puts multi-day events on every day', () => {
    const trip = ev('Trip', new Date(2026, 10, 10), new Date(2026, 10, 13), true);
    expect(eventsOnDay([trip], new Date(2026, 10, 12))).toHaveLength(1);
    expect(eventsOnDay([trip], new Date(2026, 10, 13))).toHaveLength(0);
  });

  it('validates the quick-add form', () => {
    const base = {
      calendarUrl: 'https://c/cal/home/',
      summary: 'Dentist',
      location: '',
      allDay: false,
      date: '2026-11-12',
      endDate: '',
      startTime: '09:00',
      endTime: '10:00',
    };
    expect(validateNewEvent(base)).toBeNull();
    expect(validateNewEvent({ ...base, endTime: '08:00' })).toMatch(/end after/);
    expect(validateNewEvent({ ...base, summary: ' ' })).toMatch(/title/);
    expect(validateNewEvent({ ...base, allDay: true, endDate: '2026-11-10' })).toMatch(/end date/);
    expect(calendarName('https://c/remote.php/dav/calendars/me/personal/')).toBe('personal');
  });

  it('reads the view option', () => {
    expect(parseCalendarConfig({ view: 'month' }).view).toBe('month');
    expect(parseCalendarConfig({}).view).toBe('agenda');
  });
});

describe('subscription validation', () => {
  it('requires a name and an http(s) or webcal URL for new subscriptions', () => {
    const base = { name: 'Uni', url: 'webcal://uni.example/a.ics', color: '#aa0000', refreshMinutes: 60 };
    expect(validateSubscription(base, true)).toBeNull();
    expect(validateSubscription({ ...base, name: ' ' }, true)).not.toBeNull();
    expect(validateSubscription({ ...base, url: '' }, true)).not.toBeNull();
    expect(validateSubscription({ ...base, url: '' }, false)).toBeNull();
    expect(validateSubscription({ ...base, url: 'ftp://x' }, false)).not.toBeNull();
  });
});
