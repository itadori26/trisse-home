# Meal plan

**What it does:** A week of dishes for the household, each with an optional meal, a recipe link and ingredients that go onto the shopping list with one click.

## Using it

Add the *Meal plan* widget and plan a dish with **+** next to a day. A dish
can name the meal (breakfast, lunch or dinner), link to a recipe and list its
ingredients, one per line. The arrows above the week move to the previous or
next week.

When the shopping list module is on, *Add ingredients to the shopping list*
puts the ingredients of a dish on the list the shopping widget shows. Lines
already on that list are not added twice.

The compact view shows what is planned for today.

## Who sees what

Dishes are shared with everyone on the instance by default, because a meal
plan is meant for the whole household. When you plan a dish you can narrow
it to one of your groups or to yourself. Only the person who planned a dish
can change or remove it; guests see the dishes shared with everyone when the
widget is on one of their tabs.

## Configuration

No environment variables. `MODULES_DISABLED=meals` switches the module off.

## Outbound requests

None; the module only uses the database. Recipe links are opened by the
browser, the server never fetches them.

## API

`/api/meals` (list a range of days, create, update, delete) and
`/api/meals/compact` – see `api/modules/meals.yaml`.
