import {
  type CalendarEvent,
  type CalendarSettings,
  type NewEventInput,
  validateNewEvent,
} from '../domain/calendar';

/** Upcoming events with status information. */
export interface Agenda {
  events: CalendarEvent[];
  error: string;
}

/** Output port: calendar backend. */
export interface CalendarGateway {
  settings(): Promise<CalendarSettings>;
  saveSettings(urls: string[], username: string, password: string): Promise<CalendarSettings>;
  events(days: number): Promise<Agenda>;
  range(from: Date, to: Date): Promise<Agenda>;
  create(input: {
    calendarUrl: string;
    summary: string;
    location: string;
    allDay: boolean;
    start: string;
    end: string;
  }): Promise<CalendarEvent>;
}

/** Application logic of the calendar feature. */
export class CalendarUseCases {
  constructor(private readonly gateway: CalendarGateway) {}

  settings(): Promise<CalendarSettings> {
    return this.gateway.settings();
  }

  /** Saves the connection; URLs are given one per line. */
  saveSettings(urls: string, username: string, password: string): Promise<CalendarSettings> {
    const list = urls
      .split(/\s+/)
      .map((u) => u.trim())
      .filter(Boolean);
    return this.gateway.saveSettings(list, username.trim(), password);
  }

  events(days: number): Promise<Agenda> {
    return this.gateway.events(days);
  }

  /** Events of a period, e.g. a month grid (at most 42 days). */
  range(from: Date, to: Date): Promise<Agenda> {
    return this.gateway.range(from, to);
  }

  /** Validates and creates an event from the quick-add form. */
  async createEvent(input: NewEventInput): Promise<CalendarEvent> {
    const problem = validateNewEvent(input);
    if (problem) throw new Error(problem);
    const base = {
      calendarUrl: input.calendarUrl,
      summary: input.summary.trim(),
      location: input.location.trim(),
      allDay: input.allDay,
    };
    if (input.allDay)
      return this.gateway.create({ ...base, start: input.date, end: input.endDate || input.date });
    const at = (time: string) => new Date(`${input.date}T${time}`).toISOString();
    return this.gateway.create({ ...base, start: at(input.startTime), end: at(input.endTime) });
  }
}
