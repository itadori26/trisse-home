<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { addDays, isoDate, type MealEntry, planWeek, weekStart } from '../domain/meals';
  import type { MealsUseCases, ShoppingTarget } from '../usecase/mealsUseCases';
  import MealDialog from './MealDialog.svelte';

  /** Props of the meal plan widget (#682). */
  interface Props {
    meals: MealsUseCases;
    shopping: ShoppingTarget;
  }
  let { meals, shopping }: Props = $props();
  const poll = usePolling();

  let start = $state(weekStart(new Date()));
  let entries = $state<MealEntry[]>([]);
  let error = $state('');
  let notice = $state('');
  let canShop = $state(false);
  let editing = $state<{ entry: MealEntry | null; date: string } | null>(null);

  const today = $derived(isoDate(new Date()));
  const week = $derived(planWeek(start, entries));

  async function load() {
    try {
      entries = await meals.week(start);
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => {
    void start;
    return poll(load, 15 * 60_000);
  });

  $effect(() => {
    shopping
      .available()
      .then((ok) => (canShop = ok))
      .catch(() => (canShop = false));
  });

  function dayLabel(date: string): string {
    return new Date(`${date}T12:00:00`).toLocaleDateString(intlLocale(), {
      weekday: 'short',
      day: 'numeric',
      month: 'numeric',
    });
  }

  async function remove(e: MealEntry) {
    if (!confirm(t('common.confirm_delete', { name: e.dish }))) return;
    try {
      await meals.remove(e.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function shop(e: MealEntry) {
    try {
      const n = await meals.addToShopping(e, shopping);
      notice = n === 0 ? t('meals.shopping_none') : t('meals.shopping_added', { n });
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<div class="nav">
  <button
    class="btn btn-icon"
    type="button"
    aria-label={t('meals.previous_week')}
    onclick={() => (start = addDays(start, -7))}>‹</button
  >
  <button class="btn" type="button" onclick={() => (start = weekStart(new Date()))}
    >{t('meals.this_week')}</button
  >
  <button
    class="btn btn-icon"
    type="button"
    aria-label={t('meals.next_week')}
    onclick={() => (start = addDays(start, 7))}>›</button
  >
</div>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if notice}<p class="muted" role="status">{notice}</p>{/if}
<ol class="week">
  {#each week as day (day.date)}
    <li class:today={day.date === today}>
      <div class="day">
        <span class="label">{dayLabel(day.date)}</span>
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('meals.add_on', { day: dayLabel(day.date) })}
          onclick={() => (editing = { entry: null, date: day.date })}>+</button
        >
      </div>
      {#each day.entries as e (e.id)}
        <div class="entry">
          <span class="dish">
            {#if e.meal}<span class="meal muted">{t(`meals.meal.${e.meal}`)}</span>{/if}
            {#if e.link}
              <a href={e.link} target="_blank" rel="noopener noreferrer">{e.dish}</a>
            {:else}
              {e.dish}
            {/if}
          </span>
          <span class="actions">
            {#if canShop && e.ingredients.length > 0}
              <button
                class="btn btn-icon"
                type="button"
                aria-label={t('meals.add_to_shopping', { dish: e.dish })}
                title={t('meals.add_to_shopping', { dish: e.dish })}
                onclick={() => shop(e)}>🛒</button
              >
            {/if}
            {#if e.mine}
              <button
                class="btn btn-icon"
                type="button"
                aria-label={t('meals.edit_named', { dish: e.dish })}
                onclick={() => (editing = { entry: e, date: e.date })}>✎</button
              >
              <button
                class="btn btn-icon"
                type="button"
                aria-label={t('meals.remove', { dish: e.dish })}
                onclick={() => remove(e)}>✕</button
              >
            {/if}
          </span>
        </div>
      {/each}
    </li>
  {/each}
</ol>

{#if editing}
  <MealDialog
    {meals}
    entry={editing.entry}
    date={editing.date}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = null)}
  />
{/if}

<style>
  .nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-1);
    margin-bottom: var(--space-1);
  }
  .week {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .week > li {
    border-top: 1px solid var(--color-border);
    padding-top: var(--space-1);
  }
  .today .label {
    color: var(--color-accent-text);
    font-weight: 600;
  }
  .day,
  .entry {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    min-width: 0;
  }
  .label {
    font-size: 0.85rem;
  }
  .dish {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meal {
    font-size: 0.8rem;
    margin-right: var(--space-1);
  }
  .actions {
    display: flex;
    flex-shrink: 0;
  }
</style>
