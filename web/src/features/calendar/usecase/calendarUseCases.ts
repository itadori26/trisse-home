import type { CalendarEvent, CalendarSettings } from '../domain/calendar';

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
}
