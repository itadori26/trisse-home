import type { Visibility } from '../../../shared/sharing/sharing';

/** A recurring household task (#683). */
export interface Chore {
  id: string;
  title: string;
  everyDays: number;
  /** Next due day, YYYY-MM-DD. */
  due: string;
  /** Names that take turns, in order; assignee is whose turn it is. */
  assignees: string[];
  assignee: string;
  dueNow: boolean;
  overdue: boolean;
  lastDoneAt: Date | null;
  lastDoneBy: string;
  visibility: Visibility;
  mine: boolean;
}

/** What the dialog edits; assignees one per line. */
export interface ChoreInput {
  title: string;
  everyDays: number;
  due: string;
  assignees: string;
}

/** Limits; they mirror the backend. */
export const MAX_TITLE_LENGTH = 80;
export const MAX_EVERY_DAYS = 365;
export const MAX_ASSIGNEES = 8;

/** Intervals offered as shortcuts in the dialog. */
export const INTERVAL_PRESETS = [1, 2, 3, 7, 14, 30] as const;

/** Splits the assignee text into trimmed, non-empty names. */
export function parseAssignees(text: string): string[] {
  return text
    .split(/[\n,]/)
    .map((n) => n.trim())
    .filter((n) => n !== '');
}

/** Chores due today (or overdue) and the ones still to come. */
export function splitDue(chores: Chore[]): { due: Chore[]; later: Chore[] } {
  return { due: chores.filter((c) => c.dueNow), later: chores.filter((c) => !c.dueNow) };
}

/** Days from today (local) to a YYYY-MM-DD day; negative when past. */
export function daysUntil(day: string, today: Date): number {
  const [y, m, d] = day.split('-').map(Number);
  const target = Date.UTC(y ?? 0, (m ?? 1) - 1, d ?? 1);
  const now = Date.UTC(today.getFullYear(), today.getMonth(), today.getDate());
  return Math.round((target - now) / 86_400_000);
}

/** Message keys of validateChore. */
export type ChoreError =
  'chores.error.title' | 'chores.error.every' | 'chores.error.due' | 'chores.error.assignees';

/** Validates the dialog and returns a message key or null. */
export function validateChore(input: ChoreInput): ChoreError | null {
  const title = input.title.trim();
  if (!title || [...title].length > MAX_TITLE_LENGTH) return 'chores.error.title';
  if (!Number.isInteger(input.everyDays) || input.everyDays < 1 || input.everyDays > MAX_EVERY_DAYS)
    return 'chores.error.every';
  if (!/^\d{4}-\d{2}-\d{2}$/.test(input.due)) return 'chores.error.due';
  if (parseAssignees(input.assignees).length > MAX_ASSIGNEES) return 'chores.error.assignees';
  return null;
}
