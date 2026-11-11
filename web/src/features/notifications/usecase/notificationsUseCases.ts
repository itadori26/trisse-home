import type { NotificationRules, NotificationSettings } from '../domain/notifications';

/** Input for saving the settings; an empty token keeps the stored one. */
export interface NotificationSettingsInput {
  enabled: boolean;
  serverUrl: string;
  topic: string;
  token: string;
  rules: NotificationRules;
}

/** Output port: notifications backend. */
export interface NotificationsGateway {
  settings(): Promise<NotificationSettings>;
  save(input: NotificationSettingsInput): Promise<NotificationSettings>;
  sendTest(): Promise<void>;
}

/** Application logic of the notifications feature. */
export class NotificationsUseCases {
  constructor(private readonly gateway: NotificationsGateway) {}

  settings(): Promise<NotificationSettings> {
    return this.gateway.settings();
  }

  save(input: NotificationSettingsInput): Promise<NotificationSettings> {
    return this.gateway.save({ ...input, serverUrl: input.serverUrl.trim(), topic: input.topic.trim() });
  }

  /** Saves first so the test uses what the user sees, then sends a test message. */
  async saveAndTest(input: NotificationSettingsInput): Promise<NotificationSettings> {
    const saved = await this.save(input);
    await this.gateway.sendTest();
    return saved;
  }
}
