import type { HistoryPoint, SystemSnapshot } from '../domain/system';

/** Output port: system metrics backend. */
export interface SystemGateway {
  snapshot(): Promise<SystemSnapshot>;
  history(hours: number): Promise<HistoryPoint[]>;
}

/** Application logic of the system feature. */
export class SystemUseCases {
  constructor(private readonly gateway: SystemGateway) {}

  snapshot(): Promise<SystemSnapshot> {
    return this.gateway.snapshot();
  }

  /** History of the last 24 hours. */
  history(): Promise<HistoryPoint[]> {
    return this.gateway.history(24);
  }
}
