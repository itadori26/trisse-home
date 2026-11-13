/** A habit with its statistics; streaks are computed by the backend (#106). */
export interface Habit {
  id: string;
  name: string;
  emoji: string;
  /** Weekdays 0 (Sunday) to 6; empty = every day. */
  days: number[];
  doneToday: boolean;
  current: number;
  best: number;
  /** Last 28 days, oldest first, ends today. */
  heatmap: { date: string; due: boolean; done: boolean }[];
}

/** What the user enters. */
export interface HabitInput {
  name: string;
  emoji: string;
  days: number[];
}

export const MAX_NAME_LENGTH = 60;

/** Per-widget settings. */
export interface HabitsConfig {
  showHeatmap: boolean;
}

export function parseHabitsConfig(raw: Record<string, unknown>): HabitsConfig {
  return { showHeatmap: raw['showHeatmap'] !== false };
}

/** Local calendar day as YYYY-MM-DD. */
export function isoDay(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

/** Today and yesterday – the days that can be checked off. */
export function editableDays(now: Date): { today: string; yesterday: string } {
  const y = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1);
  return { today: isoDay(now), yesterday: isoDay(y) };
}

/** Whether the habit is due on a weekday (0 = Sunday). */
export const isDue = (h: Pick<Habit, 'days'>, weekday: number): boolean =>
  h.days.length === 0 || h.days.includes(weekday);

/** Validates the input; returns a message key or null. */
export function validateHabit(i: HabitInput): 'habits.invalid_name' | null {
  const n = i.name.trim();
  return n === '' || [...n].length > MAX_NAME_LENGTH ? 'habits.invalid_name' : null;
}

/** Heatmap cells in weeks (rows of seven), oldest week first. */
export function heatWeeks<T>(cells: readonly T[]): T[][] {
  const weeks: T[][] = [];
  for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7));
  return weeks;
}
