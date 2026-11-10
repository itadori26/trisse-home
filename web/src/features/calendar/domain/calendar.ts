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

/** Configuration of the calendar widget. */
export interface CalendarConfig {
  days: number;
}

/** Reads a stored widget configuration. */
export function parseCalendarConfig(raw: Record<string, unknown>): CalendarConfig {
  const days = typeof raw['days'] === 'number' ? Math.round(raw['days']) : 7;
  return { days: Math.min(Math.max(days, 1), 31) };
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
