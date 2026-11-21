import { describe, expect, it } from 'vitest';
import {
  addDays,
  missingIngredients,
  parseIngredients,
  planWeek,
  validateMeal,
  weekStart,
  type MealEntry,
} from './meals';

const entry = (id: string, date: string): MealEntry => ({
  id,
  date,
  meal: '',
  dish: id,
  link: '',
  ingredients: [],
  visibility: 'everyone',
  mine: true,
});

describe('meal plan domain', () => {
  it('finds the Monday of a week', () => {
    expect(weekStart(new Date(2026, 10, 22))).toBe('2026-11-16'); // Sunday
    expect(weekStart(new Date(2026, 10, 23))).toBe('2026-11-23'); // Monday
    expect(addDays('2026-12-30', 3)).toBe('2027-01-02');
  });

  it('lays out a week with its dishes', () => {
    const week = planWeek('2026-11-23', [entry('a', '2026-11-24'), entry('b', '2026-12-01')]);
    expect(week).toHaveLength(7);
    expect(week[1]?.entries.map((e) => e.id)).toEqual(['a']);
    expect(week.flatMap((d) => d.entries)).toHaveLength(1);
  });

  it('adds only ingredients that are not on the list', () => {
    expect(parseIngredients(' pasta \n\n tomatoes ')).toEqual(['pasta', 'tomatoes']);
    expect(missingIngredients(['Pasta', 'Tomatoes', 'tomatoes', 'Basil'], [' pasta'])).toEqual([
      'Tomatoes',
      'Basil',
    ]);
  });

  it('validates the dialog', () => {
    const ok = { date: '2026-11-23', meal: '' as const, dish: 'Soup', link: '', ingredients: '' };
    expect(validateMeal(ok)).toBeNull();
    expect(validateMeal({ ...ok, dish: ' ' })).toBe('meals.error.dish');
    expect(validateMeal({ ...ok, link: 'javascript:alert(1)' })).toBe('meals.error.link');
    expect(validateMeal({ ...ok, date: '' })).toBe('meals.error.date');
    expect(validateMeal({ ...ok, ingredients: Array(41).fill('x').join('\n') })).toBe(
      'meals.error.ingredients',
    );
  });
});
