import { t } from '../../../shared/i18n';
/** Maps summaries containing `match` to a label, color and icon. */
export interface WasteRule {
  match: string;
  label: string;
  color: string;
  icon: string;
}

/** Settings of the waste calendar. */
export interface WasteSettings {
  configured: boolean;
  url: string;
  rules: WasteRule[];
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
