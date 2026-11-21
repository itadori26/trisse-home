import type { ShareGroup, Visibility } from '../../../shared/sharing/sharing';
import { missingIngredients, parseIngredients, type MealEntry, type MealInput } from '../domain/meals';

/** The fields the backend stores; ingredients as a list. */
export interface MealData {
  date: string;
  meal: string;
  dish: string;
  link: string;
  ingredients: string[];
}

/** Output port: meal plan backend. */
export interface MealsGateway {
  range(from: string, days: number): Promise<MealEntry[]>;
  create(data: MealData, visibility?: Visibility): Promise<MealEntry>;
  update(id: string, data: MealData, visibility?: Visibility): Promise<MealEntry>;
  remove(id: string): Promise<void>;
  shareGroups(): Promise<ShareGroup[]>;
}

/**
 * Output port: the shopping list the shopping widget shows. The app layer
 * implements it with the shopping feature, so the meal plan does not depend
 * on it and works without the shopping module.
 */
export interface ShoppingTarget {
  /** True when the shopping module is on and a list is available. */
  available(): Promise<boolean>;
  /** The open items of that list. */
  openItems(): Promise<string[]>;
  add(summary: string): Promise<void>;
}

const toData = (i: MealInput): MealData => ({
  date: i.date,
  meal: i.meal,
  dish: i.dish.trim(),
  link: i.link.trim(),
  ingredients: parseIngredients(i.ingredients),
});

/** Application logic of the meal plan (#682). */
export class MealsUseCases {
  constructor(private readonly gateway: MealsGateway) {}

  /** The dishes of the seven days from start. */
  week(start: string): Promise<MealEntry[]> {
    return this.gateway.range(start, 7);
  }

  /** Plans (no id) or changes a dish. */
  save(id: string | null, input: MealInput, visibility?: Visibility): Promise<MealEntry> {
    const data = toData(input);
    return id ? this.gateway.update(id, data, visibility) : this.gateway.create(data, visibility);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return this.gateway.shareGroups();
  }

  /**
   * Puts the ingredients of a dish on the shopping list, skipping what is
   * already open there; returns how many were added.
   */
  async addToShopping(entry: MealEntry, target: ShoppingTarget): Promise<number> {
    const missing = missingIngredients(entry.ingredients, await target.openItems());
    for (const line of missing) await target.add(line);
    return missing.length;
  }
}
