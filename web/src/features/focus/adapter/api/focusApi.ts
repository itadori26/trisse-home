import type { HttpClient } from '../../../../shared/api-client/http';
import type { WeekDTO } from '../../../../shared/api-client/generated/focus';
import type { SessionInputDTO } from '../../../../shared/api-client/generated/focus-requests';
import type { FocusDay } from '../../domain/focus';
import type { FocusGateway, Notifier } from '../../usecase/focusUseCases';

const toDays = (d: WeekDTO): FocusDay[] =>
  d.days.map((x) => ({ date: x.date, sessions: x.sessions, minutes: x.minutes }));

/** FocusGateway backed by the REST API. */
export class FocusApi implements FocusGateway {
  constructor(private readonly http: HttpClient) {}

  async stats(): Promise<FocusDay[]> {
    return toDays(await this.http.get<WeekDTO>('/api/focus/stats'));
  }

  async record(minutes: number, endedAt: Date): Promise<FocusDay[]> {
    const body: SessionInputDTO = { minutes, endedAt: endedAt.toISOString() };
    return toDays(await this.http.post<WeekDTO>('/api/focus/sessions', body));
  }
}

/** Notifier backed by the browser's Notification API. */
export const browserNotifier: Notifier = {
  permission: () => (typeof Notification === 'undefined' ? 'unsupported' : Notification.permission),
  request: async () =>
    typeof Notification === 'undefined' ? 'unsupported' : Notification.requestPermission(),
  show: (title, body) => {
    try {
      new Notification(title, { body, tag: 'trisse-focus' });
    } catch {
      // Some browsers only allow notifications from a service worker.
    }
  },
};
