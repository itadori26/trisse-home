import type { HttpClient } from '../../../../shared/api-client/http';
import type { SettingsDTO } from '../../../../shared/api-client/generated/notifications';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/notifications-requests';
import type { NotificationSettings } from '../../domain/notifications';
import type { NotificationSettingsInput, NotificationsGateway } from '../../usecase/notificationsUseCases';

const toSettings = (d: SettingsDTO): NotificationSettings => ({
  enabled: d.enabled,
  serverUrl: d.serverUrl,
  topic: d.topic,
  hasToken: d.hasToken,
  rules: { ...d.rules },
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
    const body: SettingsInputDTO = { ...input, rules: { ...input.rules } };
    return toSettings(await this.http.put<SettingsDTO>('/api/notifications/settings', body));
  }

  sendTest(): Promise<void> {
    return this.http.post('/api/notifications/test');
  }
}
