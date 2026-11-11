<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Markdown from '../../../shared/ui/Markdown.svelte';
  import type { Note } from '../domain/note';
  import type { NotesUseCases } from '../usecase/notesUseCases';

  /** Props of the note editor. */
  interface Props {
    notes: NotesUseCases;
    /** Note to edit; omitted to create one. */
    note?: Note;
    ondone: () => void;
  }
  let { notes, note, ondone }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state({ title: note?.title ?? '', body: note?.body ?? '', pinned: note?.pinned ?? true });
  let preview = $state(false);
  let error = $state('');

  async function save(e: SubmitEvent) {
    e.preventDefault();
    try {
      await notes.save(form, note?.id);
      ondone();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove() {
    if (!note || !confirm(t('common.confirm_delete', { name: note.title }))) return;
    try {
      await notes.delete(note.id);
      ondone();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={save}>
  <label class="field"
    ><span>{t('notes.title')}</span><input class="input" bind:value={form.title} maxlength="200" /></label
  >
  <div class="tabs">
    <button class="btn btn-icon" class:on={!preview} type="button" onclick={() => (preview = false)}
      >{t('notes.write')}</button
    >
    <button class="btn btn-icon" class:on={preview} type="button" onclick={() => (preview = true)}
      >{t('notes.preview')}</button
    >
  </div>
  {#if preview}
    <div class="preview"><Markdown source={form.body} /></div>
  {:else}
    <textarea
      class="input body"
      bind:value={form.body}
      rows="12"
      placeholder={t('notes.markdown_supported')}
      aria-label={t('notes.note_text')}></textarea>
  {/if}
  <label><input type="checkbox" bind:checked={form.pinned} /> {t('notes.pin_to_dashboard')}</label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p class="actions">
    <button class="btn btn-primary" type="submit">{t('notes.save')}</button>
    {#if note}<button class="btn btn-danger" type="button" onclick={remove}>{t('notes.delete')}</button>{/if}
  </p>
</form>

<style>
  .tabs {
    display: flex;
    gap: var(--space-1);
  }
  .on {
    color: var(--color-accent);
    font-weight: 600;
  }
  .body {
    font-family: var(--font-mono);
    font-size: 0.9rem;
  }
  .preview {
    min-height: 12em;
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .actions {
    display: flex;
    justify-content: space-between;
  }
</style>
