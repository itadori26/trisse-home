import { describe, expect, it, vi } from 'vitest';
import type { MealEntry } from '../domain/meals';
import { MealsUseCases, type MealsGateway, type ShoppingTarget } from './mealsUseCases';

const entry: MealEntry = {
  id: '1',
  date: '2026-11-23',
  meal: 'dinner',
  dish: 'Lasagne',
  link: '',
  ingredients: ['Pasta', 'Tomatoes'],
  visibility: 'everyone',
  mine: true,
};

function gateway(): MealsGateway {
  return {
    range: vi.fn(async () => [entry]),
    create: vi.fn(async () => entry),
    update: vi.fn(async () => entry),
    remove: vi.fn(async () => {}),
    shareGroups: vi.fn(async () => []),
  };
}

describe('MealsUseCases', () => {
  it('asks for one week and sends ingredients as a list', async () => {
    const g = gateway();
    const uc = new MealsUseCases(g);
    await uc.week('2026-11-23');
    expect(g.range).toHaveBeenCalledWith('2026-11-23', 7);
    await uc.save(null, {
      date: '2026-11-23',
      meal: '',
      dish: ' Soup ',
      link: '',
      ingredients: 'leeks\n\n potatoes ',
    });
    expect(g.create).toHaveBeenCalledWith(
      { date: '2026-11-23', meal: '', dish: 'Soup', link: '', ingredients: ['leeks', 'potatoes'] },
      undefined,
    );
    await uc.save('1', { date: '2026-11-23', meal: '', dish: 'Soup', link: '', ingredients: '' }, 'private');
    expect(g.update).toHaveBeenCalledWith('1', expect.objectContaining({ ingredients: [] }), 'private');
  });

  it('adds the missing ingredients to the shopping list', async () => {
    const added: string[] = [];
    const target: ShoppingTarget = {
      available: async () => true,
      openItems: async () => ['pasta'],
      add: async (s) => {
        added.push(s);
      },
    };
    expect(await new MealsUseCases(gateway()).addToShopping(entry, target)).toBe(1);
    expect(added).toEqual(['Tomatoes']);
  });
});
