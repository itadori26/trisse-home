import {
  type Channel,
  type ChannelInput,
  type MaintenanceWindow,
  type NotificationPolicy,
  type NotificationRules,
  type NotificationSettings,
  parseScope,
} from '../domain/notifications';

/** Input for saving the settings; an empty token keeps the stored one. */
export interface NotificationSettingsInput {
  enabled: boolean;
  serverUrl: string;
  topic: string;
  token: string;
  rules: NotificationRules;
  policy: NotificationPolicy;
}

/** Output port: notifications backend. */
export interface NotificationsGateway {
  settings(): Promise<NotificationSettings>;
  save(input: NotificationSettingsInput): Promise<NotificationSettings>;
  sendTest(): Promise<void>;
  channels(): Promise<Channel[]>;
  saveChannel(id: string | null, input: ChannelInput): Promise<Channel>;
  deleteChannel(id: string): Promise<void>;
  testChannel(id: string): Promise<void>;
  maintenance(): Promise<MaintenanceWindow[]>;
  startMaintenance(scope: string[], minutes: number, note: string): Promise<MaintenanceWindow>;
  endMaintenance(id: string): Promise<void>;
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

  channels(): Promise<Channel[]> {
    return this.gateway.channels();
  }

  saveChannel(id: string | null, input: ChannelInput): Promise<Channel> {
    return this.gateway.saveChannel(id, { ...input, name: input.name.trim() });
  }

  deleteChannel(id: string): Promise<void> {
    return this.gateway.deleteChannel(id);
  }

  testChannel(id: string): Promise<void> {
    return this.gateway.testChannel(id);
  }

  maintenance(): Promise<MaintenanceWindow[]> {
    return this.gateway.maintenance();
  }

  /** Starts a maintenance window now; scope is "service:NAS, container:nginx" or "*". */
  startMaintenance(scope: string, minutes: number, note: string): Promise<MaintenanceWindow> {
    return this.gateway.startMaintenance(parseScope(scope), Math.max(1, Math.round(minutes)), note.trim());
  }

  endMaintenance(id: string): Promise<void> {
    return this.gateway.endMaintenance(id);
  }
}
