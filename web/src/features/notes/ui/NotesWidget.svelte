<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Markdown from '../../../shared/ui/Markdown.svelte';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { excerpt, type Note, parseNotesConfig, type SearchHit } from '../domain/note';
  import type { NotesUseCases } from '../usecase/notesUseCases';
  import NoteEditor from './NoteEditor.svelte';
  import NoteHistory from './NoteHistory.svelte';
  import type { Summarizer } from '../../../shared/ui/summarizer';
  import SummaryBox from '../../../shared/ui/SummaryBox.svelte';
  import NoteViewer from './NoteViewer.svelte';

  /** Props of the notes widget. */
  interface Props {
    notes: NotesUseCases;
    config: Record<string, unknown>;
    /** Local AI summaries (#75); the button shows only when Ollama is set up. */
    summarizer?: Summarizer;
  }
  let { notes, config, summarizer }: Props = $props();
  let canSummarise = $state(false);
  $effect(() => {
    void summarizer?.available().then((ok) => (canSummarise = ok));
  });

  const cfg = $derived(parseNotesConfig(config));
  let list = $state<Note[]>([]);
  let error = $state('');
  let open = $state<Note | 'new' | null>(null);
  let viewing = $state<Note | null>(null);
  let history = $state<Note | null>(null);
  // Full-text search across all notes (#73).
  let query = $state('');
  let hits = $state<SearchHit[] | null>(null);
  let searchTimer: ReturnType<typeof setTimeout> | undefined;

  $effect(() => {
    const q = query;
    clearTimeout(searchTimer);
    if (q.trim().length < 2) {
      hits = null;
      return;
    }
    searchTimer = setTimeout(() => {
      notes
        .search(q)
        .then((h) => (hits = h))
        .catch((err) => (error = errorMessage(err)));
    }, 200);
    return () => clearTimeout(searchTimer);
  });

  async function openHit(id: string) {
    try {
      viewing = await notes.get(id);
    } catch (err) {
      error = errorMessage(err);
    }
  }
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
  <input
    class="input search"
    type="search"
    bind:value={query}
    placeholder={t('notes.search')}
    aria-label={t('notes.search_notes')}
  />
  <button class="btn btn-icon" type="button" aria-label={t('notes.new_note')} onclick={() => (open = 'new')}
    >{t('notes.note')}</button
  >
</div>
{#if error}<p class="error">{error}</p>{/if}
{#if hits}
  <ul class="list" aria-label={t('notes.search_results')}>
    {#each hits as h (h.id)}
      <li>
        <button type="button" onclick={() => openHit(h.id)}>
          <strong
            >{#each h.title as s, i (i)}{#if s.hit}<mark>{s.text}</mark>{:else}{s.text}{/if}{/each}</strong
          >
          <span class="muted"
            >{#each h.snippet as s, i (i)}{#if s.hit}<mark>{s.text}</mark>{:else}{s.text}{/if}{/each}</span
          >
        </button>
      </li>
    {:else}
      <li class="muted">{t('notes.no_matches')}</li>
    {/each}
  </ul>
{:else if single}
  {@const only = single}
  <!-- No aria-label: the visible note is the name (WCAG 2.5.3); the hidden prefix says what a click does. -->
  <button class="single" type="button" onclick={() => (open = only)}>
    <span class="sr-only">{t('notes.edit_prefix')}</span>
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
    <NoteViewer {notes} note={v} onopen={(n) => (viewing = n)} />
    {#if canSummarise && summarizer}
      {#key v.id}
        <div class="summary"><SummaryBox run={() => summarizer.text(`${v.title}\n\n${v.body}`)} /></div>
      {/key}
    {/if}
    <p class="actions">
      <button class="btn" type="button" onclick={() => ((open = v), (viewing = null))}
        >{t('notes.edit')}</button
      >
      <button class="btn" type="button" onclick={() => ((history = v), (viewing = null))}
        >{t('notes.history')}</button
      >
    </p>
  </Modal>
{/if}
{#if history}
  {@const h = history}
  <Modal title={t('notes.history_of', { title: h.title })} size="wide" onclose={() => (history = null)}>
    <NoteHistory
      {notes}
      note={h}
      onrestored={(n) => {
        history = null;
        viewing = n;
        load();
      }}
    />
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
    gap: var(--space-2);
    margin-top: -4px;
    margin-bottom: var(--space-1);
  }
  .search {
    flex: 1;
    min-width: 0;
  }
  mark {
    background: color-mix(in srgb, var(--color-accent) 30%, transparent);
    color: inherit;
    border-radius: 2px;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
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
