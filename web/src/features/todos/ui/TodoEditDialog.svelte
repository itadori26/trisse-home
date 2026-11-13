<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { MAX_REMINDER_DAYS_BEFORE, MAX_REPEAT_EVERY, type RepeatKind, type Todo } from '../domain/todo';
  import type { TodosUseCases } from '../usecase/todosUseCases';

  /** Props of the dialog that edits one to-do incl. recurrence and reminder (#68). */
  interface Props {
    todo: Todo;
    todos: TodosUseCases;
    onsaved: () => void;
    onclose: () => void;
  }
  let { todo, todos, onsaved, onclose }: Props = $props();

  const weekdays = [1, 2, 3, 4, 5, 6, 0];
  // 8 November 2026 is a Sunday, so 8 + d is weekday d.
  const dayName = (d: number) =>
    new Date(2026, 10, 8 + d).toLocaleDateString(intlLocale(), { weekday: 'short' });

  // The dialog is recreated per to-do, so copying the values once is intended.
  // svelte-ignore state_referenced_locally
  let title = $state(todo.title);
  // svelte-ignore state_referenced_locally
  let due = $state(todo.due);
  // svelte-ignore state_referenced_locally
  let kind = $state<RepeatKind>(todo.repeat.kind);
  // svelte-ignore state_referenced_locally
  let every = $state(Math.max(1, todo.repeat.every));
  // svelte-ignore state_referenced_locally
  let days = $state<number[]>([...todo.repeat.days]);
  // svelte-ignore state_referenced_locally
  let remind = $state(todo.reminder.at !== '');
  // svelte-ignore state_referenced_locally
  let at = $state(todo.reminder.at || '08:00');
  // svelte-ignore state_referenced_locally
  let daysBefore = $state(todo.reminder.daysBefore);
  let error = $state('');
  let busy = $state(false);

  const unit = $derived(
    kind === 'weekly'
      ? t('todos.unit_weeks')
      : kind === 'monthly'
        ? t('todos.unit_months')
        : t('todos.unit_days'),
  );

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    busy = true;
    try {
      await todos.editFull(todo, {
        title,
        due,
        repeat: { kind, every: Number(every), days },
        reminder: remind ? { at, daysBefore: Number(daysBefore) } : { at: '', daysBefore: 0 },
      });
      onsaved();
      onclose();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={t('todos.edit_to_do')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('todos.title')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={title} required maxlength={300} autofocus />
    </label>
    <label class="field">
      <span>{t('todos.due_date')}</span>
      <input class="input" type="date" bind:value={due} />
    </label>
    <fieldset>
      <legend>{t('todos.repeat')}</legend>
      <div class="row">
        <select class="input" bind:value={kind} aria-label={t('todos.repeat')}>
          <option value="">{t('todos.repeat_none')}</option>
          <option value="daily">{t('todos.repeat_kind_daily')}</option>
          <option value="weekly">{t('todos.repeat_kind_weekly')}</option>
          <option value="monthly">{t('todos.repeat_kind_monthly')}</option>
          <option value="after">{t('todos.repeat_kind_after')}</option>
        </select>
        {#if kind !== ''}
          <label class="inline"
            >{kind === 'after' ? t('todos.after') : t('todos.every')}
            <input class="input num" type="number" min="1" max={MAX_REPEAT_EVERY} bind:value={every} />
            {unit}</label
          >
        {/if}
      </div>
      {#if kind === 'weekly'}
        <div class="days" role="group" aria-label={t('todos.weekdays')}>
          {#each weekdays as d (d)}
            <label
              ><input
                type="checkbox"
                checked={days.includes(d)}
                onchange={(ev) =>
                  (days = ev.currentTarget.checked ? [...days, d] : days.filter((x) => x !== d))}
              />
              {dayName(d)}</label
            >
          {/each}
        </div>
      {/if}
      {#if kind !== ''}<p class="muted small">{t('todos.repeat_hint')}</p>{/if}
    </fieldset>
    <fieldset>
      <legend>{t('todos.reminder')}</legend>
      <label class="inline"
        ><input type="checkbox" bind:checked={remind} /> {t('todos.reminder_enable')}</label
      >
      {#if remind}
        <div class="row">
          <label class="inline"
            >{t('todos.at')} <input class="input" type="time" bind:value={at} required /></label
          >
          <label class="inline"
            ><input
              class="input num"
              type="number"
              min="0"
              max={MAX_REMINDER_DAYS_BEFORE}
              bind:value={daysBefore}
            />
            {t('todos.days_before')}</label
          >
        </div>
        <p class="muted small">{t('todos.reminder_hint')}</p>
      {/if}
    </fieldset>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('todos.save')}</button>
  </form>
</Modal>

<style>
  fieldset {
    margin: 0 0 0.75rem;
    display: grid;
    gap: 0.5rem;
  }
  .row,
  .days {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem 0.75rem;
    align-items: center;
  }
  .inline {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
  }
  .num {
    width: 5em;
  }
  select.input {
    width: auto;
  }
  .small {
    font-size: 0.85rem;
    margin: 0;
  }
</style>
