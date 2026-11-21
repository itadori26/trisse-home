<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import {
    type GiftIdea,
    type GiftInput,
    MAX_NOTE_LENGTH,
    MAX_PERSON_LENGTH,
    MAX_TITLE_LENGTH,
    validateGift,
  } from '../domain/gifts';
  import type { GiftsUseCases } from '../usecase/giftsUseCases';

  /** Props of the dialog that writes down or changes a gift idea (#684). */
  interface Props {
    gifts: GiftsUseCases;
    /** The idea to change, or null for a new one. */
    idea: GiftIdea | null;
    /** Prefilled person of a new idea. */
    person: string;
    /** Names to suggest: people with ideas or birthdays. */
    people: string[];
    onsaved: () => void;
    onclose: () => void;
  }
  let { gifts, idea, person, people, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<GiftInput>(
    idea
      ? { person: idea.person, title: idea.title, link: idea.link, note: idea.note, bought: idea.bought }
      : { person, title: '', link: '', note: '', bought: false },
  );
  let error = $state('');
  let busy = $state(false);
  const listId = `gift-people-${Math.random().toString(36).slice(2)}`;

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateGift(form);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await gifts.save(idea?.id ?? null, form);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={idea ? t('gifts.edit') : t('gifts.new')} {onclose}>
  <form onsubmit={submit}>
    <p class="muted small">{t('gifts.private_hint')}</p>
    <label class="field">
      <span>{t('gifts.person')}</span>
      <input class="input" bind:value={form.person} maxlength={MAX_PERSON_LENGTH} list={listId} required />
      <datalist id={listId}>
        {#each people as p (p)}<option value={p}></option>{/each}
      </datalist>
    </label>
    <label class="field">
      <span>{t('gifts.idea')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={form.title} maxlength={MAX_TITLE_LENGTH} required autofocus />
    </label>
    <label class="field">
      <span>{t('gifts.link')} {t('common.optional')}</span>
      <input class="input" type="url" bind:value={form.link} placeholder="https://" />
    </label>
    <label class="field">
      <span>{t('gifts.note')} {t('common.optional')}</span>
      <textarea class="input" rows="3" bind:value={form.note} maxlength={MAX_NOTE_LENGTH}></textarea>
    </label>
    {#if idea}
      <label class="check"><input type="checkbox" bind:checked={form.bought} /> {t('gifts.bought')}</label>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
  </form>
</Modal>

<style>
  .small {
    font-size: 0.8rem;
  }
  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 44px;
  }
</style>
