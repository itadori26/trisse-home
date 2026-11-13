<script lang="ts">
  import { palettePrefixes, QuickActionError, type QuickCommand } from '../features/quickactions';
  import { describeCommand } from './quickCapture';
  import { t } from '../shared/i18n';
  import Markdown from '../shared/ui/Markdown.svelte';
  import Modal from '../shared/ui/Modal.svelte';
  import { errorMessage } from '../shared/api-client/http';
  import type { Container } from './container';
  import { type QuickItem, search } from './quickSearch';

  /** Props of the quick search (command palette). */
  interface Props {
    container: Container;
    /** Enabled modules; disabled ones are not searched. */
    modules: Set<string>;
    onclose: () => void;
  }
  let { container, modules, onclose }: Props = $props();

  let query = $state('');
  let items = $state<QuickItem[]>([]);
  let notes = new Map<string, { title: string; body: string }>();
  let todos = new Map<string, Parameters<Container['todos']['toggle']>[0]>();
  let active = $state(0);
  let status = $state('');
  let note = $state<{ title: string; body: string } | null>(null);

  // Commands like "todo: Müll morgen" or "restart nextcloud" (#67).
  const command = $derived(container.quickActions.parse(query, palettePrefixes));
  // Body matches from the notes full-text search (#73): matched notes show
  // the excerpt around the match and count as hits for the query.
  let noteHits = $state(new Map<string, string>());
  const searchable = $derived(
    items.map((i) => {
      const snippet = i.kind === 'note' ? noteHits.get(i.id) : undefined;
      return snippet === undefined ? i : { ...i, detail: snippet, keywords: [...(i.keywords ?? []), query] };
    }),
  );
  const groups = $derived(command ? [] : search(searchable, query));
  let noteTimer: ReturnType<typeof setTimeout> | undefined;
  $effect(() => {
    const q = query;
    clearTimeout(noteTimer);
    if (!modules.has('notes') || q.trim().length < 2) {
      noteHits = new Map();
      return;
    }
    noteTimer = setTimeout(() => {
      container.notes
        .search(q)
        .then((hits) => {
          if (q === query) noteHits = new Map(hits.map((h) => [h.id, h.snippet.map((x) => x.text).join('')]));
        })
        .catch(() => (noteHits = new Map()));
    }, 200);
    return () => clearTimeout(noteTimer);
  });
  let undoLast = $state<(() => Promise<void>) | null>(null);

  async function runCommand(c: QuickCommand) {
    if (
      (c.kind === 'restart' || c.kind === 'wake') &&
      !confirm(describeCommand(c, container.quickActions) + '?')
    )
      return;
    try {
      const res = await container.quickActions.execute(c);
      query = '';
      status = describeCommand(res.command, container.quickActions, true);
      undoLast = res.undo ?? null;
    } catch (err) {
      status =
        err instanceof QuickActionError
          ? t(err.message === 'restart' ? 'quick.not_allowed_restart' : 'quick.not_allowed_wake')
          : errorMessage(err);
    }
  }

  async function undo() {
    const u = undoLast;
    undoLast = null;
    if (!u) return;
    try {
      await u();
      status = t('quick.undone');
    } catch (err) {
      status = errorMessage(err);
    }
  }
  const flat = $derived(groups.flatMap((g) => g.items));
  const optionId = (i: number) => `qs-option-${i}`;

  $effect(() => {
    // Load everything once when the palette opens; failures of one source do not block the others.
    const tasks: Promise<void>[] = [];
    const add = (list: QuickItem[]) => {
      items = [...items, ...list];
    };
    if (modules.has('bookmarks'))
      tasks.push(
        container.bookmarks.list().then((l) => {
          const folders = new Map(l.folders.map((f) => [f.id, f.name]));
          add(
            l.bookmarks.map((b) => ({
              kind: 'bookmark',
              id: b.id,
              title: b.title,
              detail: [folders.get(b.folderId), b.url].filter(Boolean).join(' · '),
              keywords: [...b.tags, b.description],
              url: b.url,
            })),
          );
        }),
      );
    if (modules.has('services'))
      tasks.push(
        container.services.list().then((l) =>
          add(
            l.map((s) => ({
              kind: 'service',
              id: s.id,
              title: s.name,
              detail: s.url,
              keywords: [s.description],
              url: s.url,
            })),
          ),
        ),
      );
    if (modules.has('notes'))
      tasks.push(
        container.notes.list('all').then((l) => {
          notes = new Map(l.map((n) => [n.id, { title: n.title, body: n.body }]));
          add(
            l.map((n) => ({
              kind: 'note',
              id: n.id,
              title: n.title || t('app.untitled_note'),
              detail: n.body.slice(0, 200),
            })),
          );
        }),
      );
    if (modules.has('todos'))
      tasks.push(
        container.todos.list(false).then((l) => {
          todos = new Map(l.map((t) => [t.id, t]));
          add(
            l.map((t) => ({ kind: 'todo', id: t.id, title: t.title, detail: t.due ? `due ${t.due}` : '' })),
          );
        }),
      );
    Promise.allSettled(tasks);
  });

  $effect(() => {
    if (!query) return; // keep the confirmation after a command cleared the field
    active = 0;
    status = '';
    undoLast = null;
  });

  async function choose(item: QuickItem | undefined) {
    if (!item) {
      const target = modules.has('search') ? container.search.targetFor(query) : null;
      if (target) window.location.assign(target);
      return;
    }
    if (item.url) {
      window.open(item.url, '_blank', 'noopener,noreferrer');
      onclose();
    } else if (item.kind === 'note') {
      note = notes.get(item.id) ?? null;
    } else if (item.kind === 'todo') {
      const todo = todos.get(item.id);
      if (!todo) return;
      try {
        await container.todos.toggle(todo);
        items = items.filter((i) => !(i.kind === 'todo' && i.id === item.id));
        status = t('app.marked_done', { title: item.title });
      } catch (err) {
        status = errorMessage(err);
      }
    }
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      // A search input would only clear itself on the first Escape.
      e.preventDefault();
      onclose();
    } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      if (flat.length) active = (active + (e.key === 'ArrowDown' ? 1 : flat.length - 1)) % flat.length;
      document.getElementById(optionId(active))?.scrollIntoView({ block: 'nearest' });
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (command) void runCommand(command);
      else choose(flat[active]);
    }
  }
</script>

<Modal title={t('app.quick_search')} {onclose}>
  {#if note}
    <div class="note">
      <button class="btn btn-icon" type="button" onclick={() => (note = null)}>{t('app.back')}</button>
      <h4>{note.title}</h4>
      <Markdown source={note.body} />
    </div>
  {:else}
    <!-- svelte-ignore a11y_autofocus -->
    <input
      class="input"
      type="search"
      role="combobox"
      aria-expanded={flat.length > 0}
      aria-controls="qs-listbox"
      aria-autocomplete="list"
      aria-activedescendant={flat.length ? optionId(active) : undefined}
      aria-label={t('app.search_bookmarks_services_notes_and_to')}
      placeholder={t('app.search_bookmarks_services_notes_to_dos')}
      autocomplete="off"
      autofocus
      bind:value={query}
      {onkeydown}
    />
    <div id="qs-listbox" role="listbox" aria-label={t('app.results')} class="results">
      {#each groups as g (g.kind)}
        <div role="group" aria-labelledby="qs-group-{g.kind}">
          <div class="group" id="qs-group-{g.kind}">{t(g.label)}</div>
          {#each g.items as it (it.kind + it.id)}
            {@const i = flat.indexOf(it)}
            <div
              id={optionId(i)}
              role="option"
              tabindex="-1"
              aria-selected={i === active}
              class="option"
              class:active={i === active}
              onclick={() => choose(it)}
              onkeydown={(e) => e.key === 'Enter' && choose(it)}
              onmousemove={() => (active = i)}
            >
              <span class="title">{it.title}</span>
              {#if it.detail}<span class="muted detail">{it.detail}</span>{/if}
            </div>
          {/each}
        </div>
      {/each}
    </div>
    {#if command}
      <p class="command" id="qs-command">
        <kbd>Enter</kbd>
        {describeCommand(command, container.quickActions)}
        <span class="muted">· <kbd>Esc</kbd> {t('quick.cancel')}</span>
      </p>
    {:else if query.trim() && flat.length === 0}
      <p class="muted hint">{t('app.no_matches_press_enter_to_search')}</p>
    {:else if !query.trim()}
      <p class="muted hint">{t('app.to_choose_enter_to_open_to')}</p>
    {/if}
    <p class="muted hint" role="status">
      {status}
      {#if undoLast}<button class="btn" type="button" onclick={undo}>{t('dashboard.undo')}</button>{/if}
    </p>
    {#if !query.trim() && !status}<p class="muted hint">{t('quick.palette_hint')}</p>{/if}
  {/if}
</Modal>

<style>
  .command {
    margin: var(--space-2) 0 0;
    padding: var(--space-2);
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
  }
  .results {
    max-height: min(60vh, 28rem);
    overflow: auto;
    margin-top: var(--space-2);
  }
  .group {
    font-size: 0.75em;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
    margin: var(--space-2) 0 2px;
  }
  .option {
    display: flex;
    flex-direction: column;
    padding: 4px 8px;
    border-radius: var(--radius-small);
    cursor: pointer;
  }
  .option.active {
    background: color-mix(in srgb, var(--color-accent) 18%, transparent);
  }
  .detail {
    font-size: 0.8em;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .hint {
    font-size: 0.8em;
    margin: var(--space-2) 0 0;
  }
  .note h4 {
    margin: var(--space-2) 0;
  }
  input {
    width: min(36rem, 80vw);
  }
</style>
