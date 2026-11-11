<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Markdown from '../../../shared/ui/Markdown.svelte';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { excerpt, type Note, parseNotesConfig } from '../domain/note';
  import type { NotesUseCases } from '../usecase/notesUseCases';
  import NoteEditor from './NoteEditor.svelte';

  /** Props of the notes widget. */
  interface Props {
    notes: NotesUseCases;
    config: Record<string, unknown>;
  }
  let { notes, config }: Props = $props();

  const cfg = $derived(parseNotesConfig(config));
  let list = $state<Note[]>([]);
  let error = $state('');
  let open = $state<Note | 'new' | null>(null);
  let viewing = $state<Note | null>(null);
  /** A single pinned note is shown in full instead of as a list. */
  const single = $derived(cfg.scope === 'pinned' && list.length === 1 ? list[0] : undefined);

  async function load() {
    try {
      list = await notes.list(cfg.scope);
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void cfg;
    load();
  });
</script>

<div class="head">
  <button class="btn btn-icon" type="button" aria-label={t('notes.new_note')} onclick={() => (open = 'new')}
    >{t('notes.note')}</button
  >
</div>
{#if error}<p class="error">{error}</p>{/if}
{#if single}
  {@const only = single}
  <button
    class="single"
    type="button"
    onclick={() => (open = only)}
    aria-label={t('notes.edit_2', { title: only.title })}
  >
    {#if !/^\s*#/.test(only.body)}<strong>{only.title}</strong>{/if}
    <Markdown source={only.body} />
  </button>
{:else}
  <ul class="list">
    {#each list as n (n.id)}
      <li>
        <button type="button" onclick={() => (viewing = n)}>
          <strong>{n.pinned ? '📌 ' : ''}{n.title}</strong>
          <span class="muted">{excerpt(n.body)}</span>
        </button>
      </li>
    {:else}
      <li class="muted">{cfg.scope === 'pinned' ? t('notes.no_pinned') : t('notes.none_yet')}</li>
    {/each}
  </ul>
{/if}

{#if viewing}
  {@const v = viewing}
  <Modal title={v.title} onclose={() => (viewing = null)}>
    <Markdown source={v.body} />
    <p>
      <button class="btn" type="button" onclick={() => ((open = v), (viewing = null))}
        >{t('notes.edit')}</button
      >
    </p>
  </Modal>
{/if}
{#if open}
  <Modal title={open === 'new' ? t('notes.new_note') : t('notes.edit_note')} onclose={() => (open = null)}>
    <NoteEditor
      {notes}
      note={open === 'new' ? undefined : open}
      ondone={() => {
        open = null;
        load();
      }}
    />
  </Modal>
{/if}

<style>
  .head {
    display: flex;
    justify-content: flex-end;
    margin-top: -4px;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .list button,
  .single {
    width: 100%;
    display: flex;
    flex-direction: column;
    text-align: left;
    border: none;
    background: var(--color-surface-2);
    border-radius: var(--radius-small);
    padding: var(--space-2);
    cursor: pointer;
  }
  .single {
    background: none;
    padding: 0;
  }
  .muted {
    font-size: 0.85em;
  }
</style>
