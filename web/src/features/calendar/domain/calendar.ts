import { t } from '../../../shared/i18n';
/** A calendar event. All-day events have date-only start/end (end exclusive). */
export interface CalendarEvent {
  uid: string;
  summary: string;
  location: string;
  /** Plain text; never rendered as HTML (#163). */
  description: string;
  /** http(s) link of the event or empty (#163). */
  url: string;
  calendar: string;
  /** Colour of an ICS subscription (#rrggbb); empty for CalDAV calendars. */
  color: string;
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

/** Layout of the calendar widget; split = month grid and events side by side (#163). */
export type CalendarView = 'agenda' | 'month' | 'split';
export const CALENDAR_VIEWS: readonly CalendarView[] = ['agenda', 'month', 'split'];

/** First day of the week: 1 = Monday (ISO, default), 0 = Sunday. */
export type WeekStart = 0 | 1;

/** Configuration of the calendar widget. */
export interface CalendarConfig {
  days: number;
  view: CalendarView;
  weekStart: WeekStart;
}

/** Reads a stored widget configuration. */
export function parseCalendarConfig(raw: Record<string, unknown>): CalendarConfig {
  const days = typeof raw['days'] === 'number' ? Math.round(raw['days']) : 7;
  const view = CALENDAR_VIEWS.includes(raw['view'] as CalendarView)
    ? (raw['view'] as CalendarView)
    : 'agenda';
  return { days: Math.min(Math.max(days, 1), 31), view, weekStart: raw['weekStart'] === 0 ? 0 : 1 };
}

/** One cell of the month grid. */
export interface MonthCell {
  key: string;
  date: Date;
  inMonth: boolean;
  isToday: boolean;
}

/**
 * Returns the 6×7 grid that shows the given month (month is 0-based, as in
 * Date); weeks start on Monday unless weekStart is 0 (Sunday).
 */
export function monthGrid(year: number, month: number, today: Date, weekStart: WeekStart = 1): MonthCell[] {
  const first = new Date(year, month, 1);
  const offset = (first.getDay() - weekStart + 7) % 7;
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

/** An ICS subscription (#98); the URL is never returned, only a hint. */
export interface CalendarSubscription {
  id: string;
  name: string;
  urlHint: string;
  color: string;
  refreshMinutes: number;
}

/** Editable fields; an empty URL keeps the stored one when updating. */
export interface SubscriptionInput {
  name: string;
  url: string;
  color: string;
  refreshMinutes: number;
}

/** Refresh intervals offered in the settings, in minutes. */
export const refreshChoices = [15, 60, 360, 1440] as const;

/** An empty subscription form. */
export const emptySubscription = (): SubscriptionInput => ({
  name: '',
  url: '',
  color: '#3b6ef5',
  refreshMinutes: 60,
});

/** Validates a subscription form; the URL is required only for new ones. */
export function validateSubscription(input: SubscriptionInput, isNew: boolean): string | null {
  if (input.name.trim() === '') return t('calendar.subscription_name_required');
  const url = input.url.trim();
  if ((isNew || url !== '') && !/^(https?|webcal):\/\/\S+$/i.test(url))
    return t('calendar.subscription_url_invalid');
  return null;
}

/** Events per day key of the cells; multi-day events on every day they cover (#163). */
export function eventsByDay(events: CalendarEvent[], cells: MonthCell[]): Map<string, CalendarEvent[]> {
  const out = new Map<string, CalendarEvent[]>();
  for (const c of cells) {
    const list = eventsOnDay(events, c.date);
    if (list.length) out.set(c.key, list);
  }
  return out;
}

/** Up to max dots for a day (colour of the calendar, '' = accent) and how many more there are. */
export function dotsForDay(events: CalendarEvent[], max = 3): { colors: string[]; more: number } {
  return { colors: events.slice(0, max).map((e) => e.color), more: Math.max(0, events.length - max) };
}

/** The next events from now on, ongoing ones included, in start order. */
export function upcoming(events: CalendarEvent[], now: Date, limit: number): CalendarEvent[] {
  return events
    .filter((e) => e.end.getTime() > now.getTime())
    .sort((a, b) => a.start.getTime() - b.start.getTime())
    .slice(0, limit);
}

/** Weekday names in the order of the grid. */
export function weekdayOrder(weekStart: WeekStart): number[] {
  return Array.from({ length: 7 }, (_, i) => (i + weekStart) % 7);
}

/**
 * Keyboard movement in the month grid (#163): arrows move by a day or a week,
 * PageUp/PageDown by a month (the day is kept where the month has it),
 * Home/End to the first/last day of the week. Null for other keys.
 */
export function moveDay(date: Date, key: string, weekStart: WeekStart = 1): Date | null {
  const y = date.getFullYear();
  const m = date.getMonth();
  const d = date.getDate();
  switch (key) {
    case 'ArrowLeft':
      return new Date(y, m, d - 1);
    case 'ArrowRight':
      return new Date(y, m, d + 1);
    case 'ArrowUp':
      return new Date(y, m, d - 7);
    case 'ArrowDown':
      return new Date(y, m, d + 7);
    case 'PageUp':
    case 'PageDown': {
      const target = m + (key === 'PageUp' ? -1 : 1);
      const last = new Date(y, target + 1, 0).getDate();
      return new Date(y, target, Math.min(d, last));
    }
    case 'Home':
      return new Date(y, m, d - ((date.getDay() - weekStart + 7) % 7));
    case 'End':
      return new Date(y, m, d + 6 - ((date.getDay() - weekStart + 7) % 7));
    default:
      return null;
  }
}

/** YYYY-MM-DD of a local date (input value of the quick-add form). */
export function dateInputValue(d: Date): string {
  return dayKey(d);
}
