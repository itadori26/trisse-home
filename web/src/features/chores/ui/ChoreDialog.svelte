<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { ShareGroup } from '../../../shared/sharing/sharing';
  import Modal from '../../../shared/ui/Modal.svelte';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';
  import {
    type Chore,
    type ChoreInput,
    INTERVAL_PRESETS,
    MAX_EVERY_DAYS,
    MAX_TITLE_LENGTH,
    validateChore,
  } from '../domain/chores';
  import type { ChoresUseCases } from '../usecase/choresUseCases';

  /** Props of the dialog that creates or changes a chore (#683). */
  interface Props {
    chores: ChoresUseCases;
    /** The chore to change, or null for a new one. */
    chore: Chore | null;
    today: string;
    onsaved: () => void;
    onclose: () => void;
  }
  let { chores, chore, today, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<ChoreInput>(
    chore
      ? {
          title: chore.title,
          everyDays: chore.everyDays,
          due: chore.due,
          assignees: chore.assignees.join('\n'),
        }
      : { title: '', everyDays: 7, due: today, assignees: '' },
  );
  // svelte-ignore state_referenced_locally
  let visibility = $state(chore?.visibility ?? 'everyone');
  let groups = $state<ShareGroup[]>([]);
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    chores
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });

  function presetLabel(days: number): string {
    if (days === 1) return t('chores.every.daily');
    if (days === 7) return t('chores.every.weekly');
    return t('chores.every.days', { n: days });
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateChore(form);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await chores.save(chore?.id ?? null, form, visibility);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={chore ? t('chores.edit') : t('chores.new')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('chores.title_label')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={form.title} maxlength={MAX_TITLE_LENGTH} required autofocus />
    </label>
    <div class="row">
      <label class="field">
        <span>{t('chores.every_label')}</span>
        <input
          class="input"
          type="number"
          min="1"
          max={MAX_EVERY_DAYS}
          bind:value={form.everyDays}
          required
        />
      </label>
      <label class="field">
        <span>{chore ? t('chores.due_label') : t('chores.first_due_label')}</span>
        <input class="input" type="date" bind:value={form.due} required />
      </label>
    </div>
    <div class="presets" role="group" aria-label={t('chores.every_label')}>
      {#each INTERVAL_PRESETS as days (days)}
        <button
          class="btn"
          type="button"
          aria-pressed={form.everyDays === days}
          onclick={() => (form.everyDays = days)}>{presetLabel(days)}</button
        >
      {/each}
    </div>
    <label class="field">
      <span>{t('chores.assignees_label')} {t('common.optional')}</span>
      <textarea class="input" rows="3" bind:value={form.assignees}></textarea>
      <span class="muted small">{t('chores.assignees_hint')}</span>
    </label>
    {#if !chore || chore.mine}
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
  .presets {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
    margin-bottom: var(--space-2);
  }
  .presets [aria-pressed='true'] {
    border-color: var(--color-accent);
  }
  .small {
    font-size: 0.8rem;
  }
</style>
