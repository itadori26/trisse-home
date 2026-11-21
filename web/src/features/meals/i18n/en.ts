/** English UI texts of the meal plan feature (#682). */
export const en = {
  'widget.meals.title': 'Meal plan',
  'widget.meals.description': 'Dishes of the week with recipe links and ingredients for the shopping list.',
  'compact.example.meals': 'Today: Lasagne',
  'meals.new': 'Plan a dish',
  'meals.edit': 'Change dish',
  'meals.dish': 'Dish',
  'meals.date': 'Day',
  'meals.meal': 'Meal',
  'meals.meal.any': 'Any time',
  'meals.meal.breakfast': 'Breakfast',
  'meals.meal.lunch': 'Lunch',
  'meals.meal.dinner': 'Dinner',
  'meals.link': 'Recipe link',
  'meals.ingredients': 'Ingredients',
  'meals.ingredients_hint': 'One per line; they can be put on the shopping list.',
  'meals.previous_week': 'Previous week',
  'meals.next_week': 'Next week',
  'meals.this_week': 'This week',
  'meals.add_on': 'Plan a dish for {day}',
  'meals.edit_named': 'Change {dish}',
  'meals.remove': 'Remove {dish}',
  'meals.add_to_shopping': 'Add the ingredients of {dish} to the shopping list',
  'meals.shopping_added': '{n} ingredients added to the shopping list.',
  'meals.shopping_none': 'All ingredients are on the shopping list already.',
  'meals.error.dish': 'Enter a dish of up to 120 characters.',
  'meals.error.date': 'Choose a day.',
  'meals.error.link': 'The link must start with http:// or https://.',
  'meals.error.ingredients': 'A dish may have at most 40 ingredients.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
