<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { ShareGroup } from '../../../shared/sharing/sharing';
  import Modal from '../../../shared/ui/Modal.svelte';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';
  import { MAX_DISH_LENGTH, MEALS, type MealEntry, type MealInput, validateMeal } from '../domain/meals';
  import type { MealsUseCases } from '../usecase/mealsUseCases';

  /** Props of the dialog that plans or changes a dish (#682). */
  interface Props {
    meals: MealsUseCases;
    /** The dish to change, or null for a new one on date. */
    entry: MealEntry | null;
    date: string;
    onsaved: () => void;
    onclose: () => void;
  }
  let { meals, entry, date, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<MealInput>(
    entry
      ? {
          date: entry.date,
          meal: entry.meal,
          dish: entry.dish,
          link: entry.link,
          ingredients: entry.ingredients.join('\n'),
        }
      : { date, meal: '', dish: '', link: '', ingredients: '' },
  );
  // svelte-ignore state_referenced_locally
  let visibility = $state(entry?.visibility ?? 'everyone');
  let groups = $state<ShareGroup[]>([]);
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    meals
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateMeal(form);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await meals.save(entry?.id ?? null, form, visibility);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={entry ? t('meals.edit') : t('meals.new')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('meals.dish')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={form.dish} maxlength={MAX_DISH_LENGTH} required autofocus />
    </label>
    <div class="row">
      <label class="field">
        <span>{t('meals.date')}</span>
        <input class="input" type="date" bind:value={form.date} required />
      </label>
      <label class="field">
        <span>{t('meals.meal')}</span>
        <select class="input" bind:value={form.meal}>
          <option value="">{t('meals.meal.any')}</option>
          {#each MEALS as m (m)}
            <option value={m}>{t(`meals.meal.${m}`)}</option>
          {/each}
        </select>
      </label>
    </div>
    <label class="field">
      <span>{t('meals.link')} {t('common.optional')}</span>
      <input class="input" type="url" bind:value={form.link} placeholder="https://" />
    </label>
    <label class="field">
      <span>{t('meals.ingredients')} {t('common.optional')}</span>
      <textarea class="input" rows="4" bind:value={form.ingredients}></textarea>
      <span class="muted small">{t('meals.ingredients_hint')}</span>
    </label>
    {#if !entry || entry.mine}
      <VisibilitySelect value={visibility} {groups} onchange={(v) => (visibility = v)} />
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
  </form>
</Modal>

<style>
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row .field {
    flex: 1 1 9rem;
  }
  .small {
    font-size: 0.8rem;
  }
</style>
