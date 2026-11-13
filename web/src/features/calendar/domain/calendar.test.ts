import { describe, expect, it } from 'vitest';
import {
  type CalendarEvent,
  calendarName,
  dotsForDay,
  eventsByDay,
  eventsOnDay,
  groupByDay,
  monthGrid,
  moveDay,
  parseCalendarConfig,
  parseLocalDate,
  upcoming,
  validateNewEvent,
  validateSubscription,
} from './calendar';

const ev = (summary: string, start: Date, end: Date, allDay = false): CalendarEvent => ({
  uid: summary,
  summary,
  location: '',
  description: '',
  url: '',
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

describe('split view domain (#163)', () => {
  const key = (d: Date) => `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()}`;
  it('starts weeks on Monday or Sunday', () => {
    // November 2026 starts on a Sunday.
    expect(monthGrid(2026, 10, new Date(2026, 10, 12), 1)[0]?.date.getDate()).toBe(26); // Mon 26 Oct
    expect(monthGrid(2026, 10, new Date(2026, 10, 12), 0)[0]?.date.getDate()).toBe(1); // Sun 1 Nov
    const cells = monthGrid(2026, 10, new Date(2026, 10, 12));
    expect(cells).toHaveLength(42);
    expect(cells.filter((c) => c.isToday).map((c) => key(c.date))).toEqual(['2026-11-12']);
    expect(cells.filter((c) => !c.inMonth).length).toBe(12);
  });
  it('puts multi-day events on every day they cover', () => {
    const cells = monthGrid(2026, 10, new Date(2026, 10, 1));
    const trip = ev('Trip', new Date(2026, 10, 10), new Date(2026, 10, 13), true); // 10–12, end exclusive
    const byDay = eventsByDay([trip], cells);
    expect([...byDay.keys()]).toEqual(['2026-11-10', '2026-11-11', '2026-11-12']);
  });
  it('gives up to three dots and the rest as a number', () => {
    const list = ['a', 'b', 'c', 'd', 'e'].map((s, i) => ({
      ...ev(s, new Date(), new Date()),
      color: i ? '' : '#f00',
    }));
    expect(dotsForDay(list)).toEqual({ colors: ['#f00', '', ''], more: 2 });
    expect(dotsForDay(list.slice(0, 2))).toEqual({ colors: ['#f00', ''], more: 0 });
  });
  it('lists upcoming events including the running one', () => {
    const now = new Date(2026, 10, 12, 12);
    const past = ev('past', new Date(2026, 10, 12, 8), new Date(2026, 10, 12, 9));
    const running = ev('running', new Date(2026, 10, 12, 11), new Date(2026, 10, 12, 13));
    const later = ev('later', new Date(2026, 10, 13, 9), new Date(2026, 10, 13, 10));
    expect(upcoming([later, past, running], now, 5).map((e) => e.summary)).toEqual(['running', 'later']);
    expect(upcoming([later, running], now, 1).map((e) => e.summary)).toEqual(['running']);
  });
  it('moves the selected day with the keyboard', () => {
    const d = new Date(2026, 10, 12); // Thursday
    const k = (key: string, ws: 0 | 1 = 1) => {
      const r = moveDay(d, key, ws);
      return r ? `${r.getFullYear()}-${r.getMonth() + 1}-${r.getDate()}` : null;
    };
    expect(k('ArrowLeft')).toBe('2026-11-11');
    expect(k('ArrowRight')).toBe('2026-11-13');
    expect(k('ArrowUp')).toBe('2026-11-5');
    expect(k('ArrowDown')).toBe('2026-11-19');
    expect(k('PageUp')).toBe('2026-10-12');
    expect(k('PageDown')).toBe('2026-12-12');
    expect(k('Home')).toBe('2026-11-9'); // Monday
    expect(k('Home', 0)).toBe('2026-11-8'); // Sunday
    expect(k('End')).toBe('2026-11-15');
    expect(k('x')).toBeNull();
    // 31 January → February keeps the last day of February.
    const jan = moveDay(new Date(2027, 0, 31), 'PageDown');
    expect(jan && `${jan.getMonth() + 1}-${jan.getDate()}`).toBe('2-28');
  });
  it('reads the split view and the week start from the config', () => {
    expect(parseCalendarConfig({ view: 'split', weekStart: 0 })).toMatchObject({
      view: 'split',
      weekStart: 0,
    });
    expect(parseCalendarConfig({})).toMatchObject({ view: 'agenda', weekStart: 1 });
    expect(parseCalendarConfig({ view: 'grid', weekStart: 3 })).toMatchObject({
      view: 'agenda',
      weekStart: 1,
    });
  });
});
