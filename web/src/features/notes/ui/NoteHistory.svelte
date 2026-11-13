<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Note, Revision, RevisionDetail } from '../domain/note';
  import type { NotesUseCases } from '../usecase/notesUseCases';

  /** Props of the version history of a note (#73). */
  interface Props {
    notes: NotesUseCases;
    note: Note;
    /** Called with the note after a version was restored. */
    onrestored: (note: Note) => void;
  }
  let { notes, note, onrestored }: Props = $props();

  let revisions = $state<Revision[] | null>(null);
  let selected = $state<RevisionDetail | null>(null);
  let error = $state('');

  const when = (d: Date) => d.toLocaleString(intlLocale(), { dateStyle: 'medium', timeStyle: 'short' });

  $effect(() => {
    const id = note.id;
    notes
      .revisions(id)
      .then((r) => (revisions = r))
      .catch((err) => (error = errorMessage(err)));
  });

  async function show(r: Revision) {
    error = '';
    try {
      selected = await notes.revision(note.id, r.id);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function restore() {
    if (!selected || !confirm(t('notes.restore_confirm', { when: when(selected.createdAt) }))) return;
    try {
      onrestored(await notes.restore(note.id, selected.id));
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if revisions && revisions.length === 0}
  <p class="muted">{t('notes.no_versions')}</p>
{:else if revisions}
  <div class="history">
    <ul class="versions" aria-label={t('notes.versions')}>
      {#each revisions as r (r.id)}
        <li>
          <button
            type="button"
            class:on={selected?.id === r.id}
            aria-current={selected?.id === r.id ? 'true' : undefined}
            onclick={() => show(r)}
          >
            <span>{when(r.createdAt)}</span>
            <span class="muted small">{r.title}</span>
          </button>
        </li>
      {/each}
    </ul>
    <div class="detail">
      {#if selected}
        <p class="muted small">{t('notes.diff_legend')}</p>
        <pre class="diff" aria-label={t('notes.changes_since')}>{#each selected.diff as line, i (i)}<span
              class="line"
              class:add={line.op === '+'}
              class:del={line.op === '-'}
              ><span class="op" aria-hidden="true">{line.op}</span>{line.text}
</span>{/each}</pre>
        <button class="btn" type="button" onclick={restore}>{t('notes.restore_version')}</button>
      {:else}
        <p class="muted">{t('notes.pick_version')}</p>
      {/if}
    </div>
  </div>
{/if}

<style>
  .history {
    display: grid;
    grid-template-columns: minmax(10rem, 14rem) 1fr;
    gap: var(--space-3);
  }
  @media (max-width: 600px) {
    .history {
      grid-template-columns: 1fr;
    }
  }
  .versions {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
    max-height: 50vh;
    overflow: auto;
  }
  .versions button {
    width: 100%;
    display: flex;
    flex-direction: column;
    text-align: left;
    border: 1px solid var(--color-border);
    background: var(--color-surface);
    border-radius: var(--radius-small);
    padding: var(--space-1) var(--space-2);
    cursor: pointer;
  }
  .versions button.on {
    border-color: var(--color-accent);
  }
  .diff {
    max-height: 45vh;
    overflow: auto;
    font-family: var(--font-mono);
    font-size: 0.85rem;
    background: var(--color-surface-2);
    border-radius: var(--radius-small);
    padding: var(--space-2);
    white-space: pre-wrap;
  }
  .line {
    display: block;
  }
  .op {
    display: inline-block;
    width: 1.2em;
    user-select: none;
  }
  .add {
    background: color-mix(in srgb, var(--color-success, #2e7d32) 18%, transparent);
  }
  .del {
    background: color-mix(in srgb, var(--color-danger) 18%, transparent);
    text-decoration: line-through;
  }
  .small {
    font-size: 0.8rem;
  }
</style>
