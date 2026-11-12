import type { EventSeverity, TimelineEvent } from '../domain/timeline';

/** Filter of the timeline list; empty values match everything. */
export interface TimelineFilter {
  source: string;
  severity: EventSeverity | '';
  limit: number;
}

/** Output port: timeline backend. */
export interface TimelineGateway {
  list(filter: TimelineFilter): Promise<TimelineEvent[]>;
  retentionDays(): Promise<number>;
  setRetentionDays(days: number): Promise<number>;
}

/** Application logic of the timeline feature. */
export class TimelineUseCases {
  constructor(private readonly gateway: TimelineGateway) {}

  list(filter: TimelineFilter): Promise<TimelineEvent[]> {
    return this.gateway.list(filter);
  }

  retentionDays(): Promise<number> {
    return this.gateway.retentionDays();
  }

  setRetentionDays(days: number): Promise<number> {
    return this.gateway.setRetentionDays(Math.round(days));
  }
}
