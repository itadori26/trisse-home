import type { HttpClient } from '../../../../shared/api-client/http';
import type { AgendaDTO, EventDTO, SettingsDTO } from '../../../../shared/api-client/generated/calendar';
import type {
  NewEventDTO,
  SettingsInputDTO,
} from '../../../../shared/api-client/generated/calendar-requests';
import { type CalendarEvent, type CalendarSettings, parseLocalDate } from '../../domain/calendar';
import type { Agenda, CalendarGateway } from '../../usecase/calendarUseCases';

const toEvent = (d: EventDTO): CalendarEvent => ({
  uid: d.uid,
  summary: d.summary,
  location: d.location,
  calendar: d.calendar,
  allDay: d.allDay,
  // All-day dates are calendar dates, not instants: interpret them locally.
  start: d.allDay ? parseLocalDate(d.start) : new Date(d.start),
  end: d.allDay ? parseLocalDate(d.end) : new Date(d.end),
});

/** CalendarGateway backed by the REST API. */
export class CalendarApi implements CalendarGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<CalendarSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/calendar/settings')) };
  }
  async saveSettings(urls: string[], username: string, password: string): Promise<CalendarSettings> {
    const body: SettingsInputDTO = { calendarUrls: urls, username, password };
    return { ...(await this.http.put<SettingsDTO>('/api/calendar/settings', body)) };
  }
  async events(days: number): Promise<Agenda> {
    const d = await this.http.get<AgendaDTO>(`/api/calendar/events?days=${days}`);
    return { events: d.events.map(toEvent), error: d.error };
  }
  async range(from: Date, to: Date): Promise<Agenda> {
    const q = new URLSearchParams({ from: from.toISOString(), to: to.toISOString() });
    const d = await this.http.get<AgendaDTO>(`/api/calendar/events?${q}`);
    return { events: d.events.map(toEvent), error: d.error };
  }
  async create(input: NewEventDTO): Promise<CalendarEvent> {
    return toEvent(await this.http.post<EventDTO>('/api/calendar/events', input));
  }
}
