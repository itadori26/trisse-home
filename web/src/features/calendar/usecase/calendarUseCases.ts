import {
  type CalendarEvent,
  type CalendarSettings,
  type CalendarSubscription,
  type NewEventInput,
  type SubscriptionInput,
  validateNewEvent,
  validateSubscription,
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
  subscriptions(): Promise<CalendarSubscription[]>;
  addSubscription(input: SubscriptionInput): Promise<CalendarSubscription>;
  updateSubscription(id: string, input: SubscriptionInput): Promise<CalendarSubscription>;
  deleteSubscription(id: string): Promise<void>;
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

  /** ICS subscriptions (#98). */
  subscriptions(): Promise<CalendarSubscription[]> {
    return this.gateway.subscriptions();
  }

  /** Adds or updates a subscription after validation; an empty URL keeps the stored one. */
  saveSubscription(input: SubscriptionInput, id?: string): Promise<CalendarSubscription> {
    const problem = validateSubscription(input, !id);
    if (problem) return Promise.reject(new Error(problem));
    const clean = { ...input, name: input.name.trim(), url: input.url.trim() };
    return id ? this.gateway.updateSubscription(id, clean) : this.gateway.addSubscription(clean);
  }

  deleteSubscription(id: string): Promise<void> {
    return this.gateway.deleteSubscription(id);
  }
}
