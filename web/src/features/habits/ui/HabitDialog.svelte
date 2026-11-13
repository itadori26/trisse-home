<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { type Habit, type HabitInput, MAX_NAME_LENGTH, validateHabit } from '../domain/habits';
  import type { HabitsUseCases } from '../usecase/habitsUseCases';

  /** Props of the dialog that creates or edits a habit (#106). */
  interface Props {
    habits: HabitsUseCases;
    /** The habit to edit, or null for a new one. */
    item: Habit | null;
    onsaved: () => void;
    ondelete?: () => void;
    onclose: () => void;
  }
  let { habits, item, onsaved, ondelete, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<HabitInput>(
    item ? { name: item.name, emoji: item.emoji, days: [...item.days] } : { name: '', emoji: '', days: [] },
  );
  let error = $state('');
  let busy = $state(false);

  // Monday first, as in German calendars; values stay 0 = Sunday.
  const weekdays = [1, 2, 3, 4, 5, 6, 0];
  const dayName = (d: number) =>
    new Date(2026, 10, 15 + d).toLocaleDateString(intlLocale(), { weekday: 'short' }); // 15 Nov 2026 is a Sunday

  function toggleDay(d: number, on: boolean) {
    form.days = on ? [...form.days, d] : form.days.filter((x) => x !== d);
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateHabit(form);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await habits.save(item?.id ?? null, form);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={item ? t('habits.edit') : t('habits.new')} {onclose}>
  <form onsubmit={submit}>
    <div class="row">
      <label class="field grow">
        <span>{t('habits.name')}</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input class="input" bind:value={form.name} maxlength={MAX_NAME_LENGTH} required autofocus />
      </label>
      <label class="field emoji">
        <span>{t('habits.emoji')}</span>
        <input class="input" bind:value={form.emoji} maxlength="8" placeholder="💧" />
      </label>
    </div>
    <fieldset>
      <legend>{t('habits.schedule')}</legend>
      <p class="muted small">{t('habits.schedule_hint')}</p>
      <div class="days">
        {#each weekdays as d (d)}
          <label
            ><input
              type="checkbox"
              checked={form.days.includes(d)}
              onchange={(e) => toggleDay(d, e.currentTarget.checked)}
            />
            {dayName(d)}</label
          >
        {/each}
      </div>
    </fieldset>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <p class="actions">
      <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
      {#if item && ondelete}
        <button class="btn" type="button" onclick={ondelete}>{t('habits.delete')}</button>
      {/if}
    </p>
  </form>
</Modal>

<style>
  .row {
    display: flex;
    gap: var(--space-2);
  }
  .grow {
    flex: 1;
  }
  .emoji {
    flex: 0 1 5rem;
  }
  fieldset {
    border: none;
    padding: 0;
    margin: var(--space-2) 0;
  }
  legend {
    font-weight: 500;
  }
  .days {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .small {
    font-size: 0.8rem;
    margin: 0 0 var(--space-1);
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
</style>
