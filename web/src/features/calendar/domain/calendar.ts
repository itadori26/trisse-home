import { t } from '../../../shared/i18n';
/** A calendar event. All-day events have date-only start/end (end exclusive). */
export interface CalendarEvent {
  uid: string;
  summary: string;
  location: string;
  calendar: string;
  start: Date;
  end: Date;
  allDay: boolean;
}

/** Events of one day. */
export interface AgendaDay {
  /** Local date key YYYY-MM-DD. */
  key: string;
  date: Date;
  events: CalendarEvent[];
}

/** Connection settings as shown in the UI (the password is never returned). */
export interface CalendarSettings {
  configured: boolean;
  calendarUrls: string[];
  username: string;
  hasPassword: boolean;
}

/** Layout of the calendar widget. */
export type CalendarView = 'agenda' | 'month';

/** Configuration of the calendar widget. */
export interface CalendarConfig {
  days: number;
  view: CalendarView;
}

/** Reads a stored widget configuration. */
export function parseCalendarConfig(raw: Record<string, unknown>): CalendarConfig {
  const days = typeof raw['days'] === 'number' ? Math.round(raw['days']) : 7;
  return { days: Math.min(Math.max(days, 1), 31), view: raw['view'] === 'month' ? 'month' : 'agenda' };
}

/** One cell of the month grid. */
export interface MonthCell {
  key: string;
  date: Date;
  inMonth: boolean;
  isToday: boolean;
}

/**
 * Returns the 6×7 grid (weeks start on Monday) that shows the given month
 * (month is 0-based, as in Date).
 */
export function monthGrid(year: number, month: number, today: Date): MonthCell[] {
  const first = new Date(year, month, 1);
  const offset = (first.getDay() + 6) % 7; // Monday = 0
  const todayKey = dayKey(today);
  const cells: MonthCell[] = [];
  for (let i = 0; i < 42; i++) {
    const date = new Date(year, month, 1 - offset + i);
    const key = dayKey(date);
    cells.push({ key, date, inMonth: date.getMonth() === month, isToday: key === todayKey });
  }
  return cells;
}

/** Events overlapping the local day of date (multi-day events on every day). */
export function eventsOnDay(events: CalendarEvent[], date: Date): CalendarEvent[] {
  const dayStart = new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
  const dayEnd = new Date(date.getFullYear(), date.getMonth(), date.getDate() + 1).getTime();
  return events.filter((e) => e.start.getTime() < dayEnd && e.end.getTime() > dayStart);
}

/** Input of the quick-add form. */
export interface NewEventInput {
  calendarUrl: string;
  summary: string;
  location: string;
  allDay: boolean;
  /** YYYY-MM-DD */
  date: string;
  /** YYYY-MM-DD, inclusive; only for all-day events */
  endDate: string;
  /** HH:MM; only for timed events */
  startTime: string;
  endTime: string;
}

/** Validates the quick-add form and returns an error message or null. */
export function validateNewEvent(i: NewEventInput): string | null {
  if (!i.calendarUrl) return t('calendar.choose_calendar');
  if (i.summary.trim() === '') return t('calendar.enter_title');
  if (!/^\d{4}-\d{2}-\d{2}$/.test(i.date)) return t('calendar.enter_date');
  if (i.allDay) {
    if (i.endDate && i.endDate < i.date) return t('calendar.end_before_start_date');
    return null;
  }
  if (!/^\d{2}:\d{2}$/.test(i.startTime) || !/^\d{2}:\d{2}$/.test(i.endTime))
    return t('calendar.enter_times');
  if (i.endTime <= i.startTime) return t('calendar.end_after_start');
  return null;
}

/** The display name of a calendar URL (last path segment). */
export function calendarName(url: string): string {
  const parts = url.replace(/\/+$/, '').split('/');
  return decodeURIComponent(parts[parts.length - 1] ?? url);
}

const pad = (n: number) => String(n).padStart(2, '0');
const dayKey = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;

/**
 * Groups events by local day. Multi-day events appear on every day they
 * cover, starting today at the earliest.
 */
export function groupByDay(events: CalendarEvent[], today: Date, days: number): AgendaDay[] {
  const start = new Date(today.getFullYear(), today.getMonth(), today.getDate());
  const map = new Map<string, AgendaDay>();
  for (let i = 0; i < days; i++) {
    const date = new Date(start.getFullYear(), start.getMonth(), start.getDate() + i);
    const dayStart = date.getTime();
    const dayEnd = new Date(date.getFullYear(), date.getMonth(), date.getDate() + 1).getTime();
    for (const e of events) {
      if (e.start.getTime() < dayEnd && e.end.getTime() > dayStart) {
        const key = dayKey(date);
        const entry = map.get(key) ?? { key, date, events: [] };
        entry.events.push(e);
        map.set(key, entry);
      }
    }
  }
  return [...map.values()];
}

/** Parses a date-only string as a local date. */
export function parseLocalDate(s: string): Date {
  const [y, m, d] = s.split('-').map(Number);
  return new Date(y ?? 1970, (m ?? 1) - 1, d ?? 1);
}
