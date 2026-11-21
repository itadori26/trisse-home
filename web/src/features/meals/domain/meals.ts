import type { Visibility } from '../../../shared/sharing/sharing';

/** The optional meal of a dish, in the order of a day. */
export const MEALS = ['breakfast', 'lunch', 'dinner'] as const;
export type Meal = (typeof MEALS)[number] | '';

/** A planned dish (#682). */
export interface MealEntry {
  id: string;
  /** YYYY-MM-DD */
  date: string;
  meal: Meal;
  dish: string;
  link: string;
  ingredients: string[];
  visibility: Visibility;
  /** True when the viewer planned the dish and may change it. */
  mine: boolean;
}

/** What the dialog edits. Ingredients are one per line. */
export interface MealInput {
  date: string;
  meal: Meal;
  dish: string;
  link: string;
  ingredients: string;
}

/** Limits; they mirror the backend. */
export const MAX_DISH_LENGTH = 120;
export const MAX_INGREDIENTS = 40;

/** A day of the week view with its dishes. */
export interface PlanDay {
  date: string;
  entries: MealEntry[];
}

/** Local date as YYYY-MM-DD. */
export function isoDate(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

/** The day n days after a YYYY-MM-DD date. Noon avoids daylight saving edges. */
export function addDays(date: string, n: number): string {
  const d = new Date(`${date}T12:00:00`);
  d.setDate(d.getDate() + n);
  return isoDate(d);
}

/** The Monday of the week that contains d. */
export function weekStart(d: Date): string {
  const day = (d.getDay() + 6) % 7; // Monday = 0
  return addDays(isoDate(d), -day);
}

/** Seven days from start, each with its dishes in the order they came. */
export function planWeek(start: string, entries: MealEntry[]): PlanDay[] {
  return Array.from({ length: 7 }, (_, i) => {
    const date = addDays(start, i);
    return { date, entries: entries.filter((e) => e.date === date) };
  });
}

/** Splits the ingredient text into trimmed, non-empty lines. */
export function parseIngredients(text: string): string[] {
  return text
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l !== '');
}

/**
 * The ingredients that are not on the shopping list yet, compared without
 * case and surrounding space, each once.
 */
export function missingIngredients(ingredients: string[], onList: string[]): string[] {
  const seen = new Set(onList.map((s) => s.trim().toLowerCase()));
  const out: string[] = [];
  for (const i of ingredients) {
    const key = i.trim().toLowerCase();
    if (key && !seen.has(key)) {
      seen.add(key);
      out.push(i.trim());
    }
  }
  return out;
}

/** Message keys of validateMeal. */
export type MealError =
  'meals.error.dish' | 'meals.error.date' | 'meals.error.link' | 'meals.error.ingredients';

/** Validates the dialog and returns a message key or null. */
export function validateMeal(input: MealInput): MealError | null {
  const dish = input.dish.trim();
  if (!dish) return 'meals.error.dish';
  if ([...dish].length > MAX_DISH_LENGTH) return 'meals.error.dish';
  if (!/^\d{4}-\d{2}-\d{2}$/.test(input.date)) return 'meals.error.date';
  const link = input.link.trim();
  if (link && !/^https?:\/\/[^\s/]+/i.test(link)) return 'meals.error.link';
  if (parseIngredients(input.ingredients).length > MAX_INGREDIENTS) return 'meals.error.ingredients';
  return null;
}
