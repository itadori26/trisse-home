<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { type Countdown, type CountdownInput, MAX_TITLE_LENGTH, validateInput } from '../domain/countdown';
  import type { CountdownUseCases } from '../usecase/countdownUseCases';

  /** Props of the dialog that creates or edits a countdown (#54). */
  interface Props {
    countdown: CountdownUseCases;
    /** The countdown to edit, or null for a new one. */
    item: Countdown | null;
    onsaved: () => void;
    onclose: () => void;
  }
  let { countdown, item, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<CountdownInput>(
    item
      ? { title: item.title, emoji: item.emoji, date: item.date, time: item.time }
      : { title: '', emoji: '', date: '', time: '' },
  );
  let error = $state('');
  let busy = $state(false);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateInput(form);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await countdown.save(item?.id ?? null, form);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={item ? t('countdown.edit') : t('countdown.new')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('countdown.title')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={form.title} maxlength={MAX_TITLE_LENGTH} required autofocus />
    </label>
    <div class="row">
      <label class="field">
        <span>{t('countdown.date')}</span>
        <input class="input" type="date" bind:value={form.date} required />
      </label>
      <label class="field">
        <span>{t('countdown.time')} {t('common.optional')}</span>
        <input class="input" type="time" bind:value={form.time} />
      </label>
      <label class="field emoji">
        <span>{t('countdown.emoji')}</span>
        <input class="input" bind:value={form.emoji} maxlength="8" placeholder="🎉" />
      </label>
    </div>
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
    flex: 1 1 8rem;
  }
  .row .emoji {
    flex: 0 1 5rem;
  }
</style>
