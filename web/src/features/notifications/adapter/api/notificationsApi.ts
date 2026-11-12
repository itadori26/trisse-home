import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  ChannelViewDTO,
  MaintenanceWindowDTO,
  SettingsDTO,
} from '../../../../shared/api-client/generated/notifications';
import type {
  ChannelDTO,
  MaintenanceDTO,
  SettingsInputDTO,
} from '../../../../shared/api-client/generated/notifications-requests';
import type {
  Channel,
  ChannelInput,
  MaintenanceWindow,
  NotificationSettings,
  Severity,
} from '../../domain/notifications';
import type { NotificationSettingsInput, NotificationsGateway } from '../../usecase/notificationsUseCases';

const toSettings = (d: SettingsDTO): NotificationSettings => ({
  enabled: d.enabled,
  serverUrl: d.serverUrl,
  topic: d.topic,
  hasToken: d.hasToken,
  rules: { ...d.rules },
  policy: {
    severities: { ...d.policy.severities } as Record<string, Severity>,
    quiet: { ...d.policy.quiet, days: d.policy.quiet.days ?? [] },
  },
  lastError: d.lastError,
  lastDelivered: d.lastDelivered ? new Date(d.lastDelivered) : null,
});

/** NotificationsGateway backed by the REST API. */
export class NotificationsApi implements NotificationsGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<NotificationSettings> {
    return toSettings(await this.http.get<SettingsDTO>('/api/notifications/settings'));
  }

  async save(input: NotificationSettingsInput): Promise<NotificationSettings> {
    const body: SettingsInputDTO = { ...input, rules: { ...input.rules }, policy: input.policy };
    return toSettings(await this.http.put<SettingsDTO>('/api/notifications/settings', body));
  }

  sendTest(): Promise<void> {
    return this.http.post('/api/notifications/test');
  }

  async channels(): Promise<Channel[]> {
    return (await this.http.get<ChannelViewDTO[]>('/api/notifications/channels')).map(toChannel);
  }

  async saveChannel(id: string | null, input: ChannelInput): Promise<Channel> {
    const body: ChannelDTO = { ...input };
    const res = id
      ? await this.http.put<ChannelViewDTO>(`/api/notifications/channels/${encodeURIComponent(id)}`, body)
      : await this.http.post<ChannelViewDTO>('/api/notifications/channels', body);
    return toChannel(res);
  }

  deleteChannel(id: string): Promise<void> {
    return this.http.delete(`/api/notifications/channels/${encodeURIComponent(id)}`);
  }

  testChannel(id: string): Promise<void> {
    return this.http.post(`/api/notifications/channels/${encodeURIComponent(id)}/test`);
  }

  async maintenance(): Promise<MaintenanceWindow[]> {
    return (await this.http.get<MaintenanceWindowDTO[]>('/api/notifications/maintenance')).map(toWindow);
  }

  async startMaintenance(scope: string[], minutes: number, note: string): Promise<MaintenanceWindow> {
    const body: MaintenanceDTO = { scope, minutes, note };
    return toWindow(await this.http.post<MaintenanceWindowDTO>('/api/notifications/maintenance', body));
  }

  endMaintenance(id: string): Promise<void> {
    return this.http.delete(`/api/notifications/maintenance/${encodeURIComponent(id)}`);
  }
}

const toChannel = (d: ChannelViewDTO): Channel => ({ ...d, types: d.types ?? [] }) as Channel;

const toWindow = (d: MaintenanceWindowDTO): MaintenanceWindow => ({
  id: d.id,
  scope: d.scope ?? [],
  start: new Date(d.start),
  end: new Date(d.end),
  note: d.note,
});
