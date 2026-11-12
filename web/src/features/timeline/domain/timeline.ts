/** Importance of a timeline event. */
export type EventSeverity = 'info' | 'warning' | 'critical';

/** One entry of the event timeline (#74). */
export interface TimelineEvent {
  id: string;
  at: Date;
  source: string;
  severity: EventSeverity;
  subject: string;
  message: string;
}

/** Sources that can be filtered. */
export const eventSources = [
  'service',
  'container',
  'backup',
  'check',
  'tls',
  'disk',
  'diskhealth',
  'security',
] as const;

/** A day heading followed by its events. */
export interface DayGroup {
  day: string;
  events: TimelineEvent[];
}

/** Groups events (newest first) by local calendar day for day separators. */
export function groupByDay(events: readonly TimelineEvent[], dayKey: (d: Date) => string): DayGroup[] {
  const groups: DayGroup[] = [];
  for (const e of events) {
    const day = dayKey(e.at);
    const last = groups.at(-1);
    if (last && last.day === day) last.events.push(e);
    else groups.push({ day, events: [e] });
  }
  return groups;
}

/** Widget options. */
export interface TimelineConfig {
  /** Compact mode shows only the last N events without filters. */
  compact: boolean;
  limit: number;
}

/** Reads the widget configuration with defaults. */
export function timelineConfig(raw: Record<string, unknown>): TimelineConfig {
  const limit = typeof raw['limit'] === 'number' ? raw['limit'] : 10;
  return { compact: raw['compact'] !== false, limit: Math.min(100, Math.max(3, Math.round(limit))) };
}
