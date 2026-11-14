import { t } from '../../../shared/i18n';
/** Maps summaries containing `match` to a label, color and icon. */
export interface WasteRule {
  match: string;
  label: string;
  color: string;
  icon: string;
}

/** Where the dates come from (#153); '' = not set up. */
export type WasteSource = 'url' | 'file' | '';

/** An uploaded ICS file; the content stays on the server. */
export interface CalendarFile {
  name: string;
  uploadedAt: Date;
  events: number;
  /** First and last collection date. */
  from: Date;
  to: Date;
}

/** Settings of the waste calendar. */
export interface WasteSettings {
  configured: boolean;
  source: WasteSource;
  url: string;
  file: CalendarFile | null;
  rules: WasteRule[];
}

/** Largest ICS file the server accepts (1 MB). */
export const MAX_CALENDAR_BYTES = 1 << 20;

/**
 * Checks a chosen file before it is uploaded (the server checks again):
 * an .ics file or text/calendar of at most 1 MB. Returns an error or null.
 */
export function validateCalendarFile(f: { name: string; size: number; type: string }): string | null {
  if (!/\.ics$/i.test(f.name) && !f.type.startsWith('text/calendar')) return t('waste.file_not_ics');
  if (f.size > MAX_CALENDAR_BYTES) return t('waste.file_too_large');
  if (f.size === 0) return t('waste.file_empty');
  return null;
}

/** One waste type with its next collection dates. */
export interface WasteType {
  label: string;
  color: string;
  icon: string;
  dates: Date[];
  daysUntil: number;
}

/** Overview of the next collections. */
export interface WasteOverview {
  types: WasteType[];
  error: string;
  /** Last date of an uploaded calendar that ends within 30 days, else null (#153). */
  endsOn: Date | null;
}

/** "Today", "Tomorrow" or "in 3 days". */
export function dueLabel(daysUntil: number): string {
  if (daysUntil <= 0) return t('time.today');
  if (daysUntil === 1) return t('time.tomorrow');
  return t('time.in_days', { n: daysUntil });
}

/** Whether a collection needs attention: tomorrow, or today before noon. */
export function isDueSoon(t: Pick<WasteType, 'daysUntil'>, now: Date): boolean {
  return t.daysUntil === 1 || (t.daysUntil === 0 && now.getHours() < 12);
}

/** An empty rule for the editor. */
export const emptyRule = (): WasteRule => ({ match: '', label: '', color: '', icon: '' });
