import { resolveCurrent } from '../../features/shopping/domain/shopping';
import type { ShoppingTarget } from '../../features/meals/usecase/mealsUseCases';
import { t } from '../../shared/i18n';
import type { Container } from '../container';
import { lazy, type WidgetDefinition } from '../widgetKit';

/**
 * The shopping list the shopping widget shows, as the meal plan's port: the
 * features stay independent, and without the shopping module the button is
 * not offered.
 */
function shoppingTarget(c: Container): ShoppingTarget {
  const current = async () => resolveCurrent(await c.shopping.overview());
  return {
    async available() {
      return (await c.enabledModules()).modules.has('shopping') && (await current()) !== '';
    },
    async openItems() {
      const items = await c.shopping.source(await current()).items();
      return items.filter((i) => !i.done).map((i) => i.summary);
    },
    async add(summary) {
      await c.shopping.source(await current()).add(summary);
    },
  };
}

/** The meal plan widget (#682). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'meals',
      category: 'everyday',
      keywords: [
        'meal',
        'dinner',
        'recipe',
        'menu',
        'family',
        'essen',
        'essensplan',
        'rezept',
        'familie',
        'kochen',
      ],
      get title() {
        return t('widget.meals.title');
      },
      get description() {
        return t('widget.meals.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'calendar',
    compact: () => 'meals',
    compactExample: () => t('compact.example.meals'),
    view: lazy(
      () => import('../../features/meals/ui/MealsWidget.svelte'),
      (_config, c) => ({ meals: c.meals, shopping: shoppingTarget(c) }),
    ),
  },
];
