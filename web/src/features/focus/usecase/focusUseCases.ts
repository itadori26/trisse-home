import type { FocusDay } from '../domain/focus';

/** Output port: focus backend. */
export interface FocusGateway {
  stats(): Promise<FocusDay[]>;
  /** Records a finished session and returns the updated statistics. */
  record(minutes: number, endedAt: Date): Promise<FocusDay[]>;
}

/** Output port: browser notifications. */
export interface Notifier {
  permission(): NotificationPermission | 'unsupported';
  request(): Promise<NotificationPermission | 'unsupported'>;
  show(title: string, body: string): void;
}

/** Application logic of the focus timer (#58). */
export class FocusUseCases {
  constructor(
    private readonly gateway: FocusGateway,
    readonly notifier: Notifier,
  ) {}

  stats(): Promise<FocusDay[]> {
    return this.gateway.stats();
  }

  /** Records a finished focus session; breaks are not recorded. */
  record(minutes: number, endedAt: Date): Promise<FocusDay[]> {
    return this.gateway.record(minutes, endedAt);
  }

  /** Notifies about the end of a phase if enabled and allowed. */
  notify(enabled: boolean, title: string, body: string): void {
    if (enabled && this.notifier.permission() === 'granted') this.notifier.show(title, body);
  }
}
