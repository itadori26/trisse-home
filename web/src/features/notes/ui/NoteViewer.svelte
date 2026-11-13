<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Markdown from '../../../shared/ui/Markdown.svelte';
  import { findByTitle, linkWikiLinks, type Note, type NoteRef, noteLinkTitle } from '../domain/note';
  import type { NotesUseCases } from '../usecase/notesUseCases';

  /** Props of the read view of a note with wiki links and backlinks (#73). */
  interface Props {
    notes: NotesUseCases;
    note: Note;
    /** Opens another note (followed link or backlink). */
    onopen: (note: Note) => void;
  }
  let { notes, note, onopen }: Props = $props();

  let backlinks = $state<NoteRef[]>([]);
  let error = $state('');

  $effect(() => {
    const id = note.id;
    notes
      .backlinks(id)
      .then((b) => (backlinks = b))
      .catch(() => (backlinks = []));
  });

  async function follow(href: string) {
    const title = noteLinkTitle(href);
    if (title === null) return;
    error = '';
    try {
      const target = findByTitle(await notes.list('all'), title);
      if (target) onopen(target);
      else error = t('notes.link_target_missing', { title });
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function openRef(ref: NoteRef) {
    try {
      onopen(await notes.get(ref.id));
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<Markdown source={linkWikiLinks(note.body)} onlink={follow} />
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if backlinks.length > 0}
  <section class="backlinks" aria-label={t('notes.backlinks')}>
    <h4>{t('notes.backlinks')}</h4>
    <ul>
      {#each backlinks as b (b.id)}
        <li><button class="link" type="button" onclick={() => openRef(b)}>{b.title}</button></li>
      {/each}
    </ul>
  </section>
{/if}

<style>
  .backlinks {
    margin-top: var(--space-3);
    border-top: 1px solid var(--color-border);
    padding-top: var(--space-2);
  }
  h4 {
    margin: 0 0 var(--space-1);
    font-size: 0.85rem;
    color: var(--color-text-muted);
  }
  ul {
    margin: 0;
    padding-left: 1.2em;
  }
  .link {
    border: none;
    background: none;
    padding: 0;
    color: var(--color-accent);
    cursor: pointer;
    text-align: left;
  }
  .link:hover {
    text-decoration: underline;
  }
</style>
