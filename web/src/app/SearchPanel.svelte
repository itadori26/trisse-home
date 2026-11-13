<script lang="ts">
  import {
    captureRest,
    mergePrefixes,
    palettePrefixes,
    prefixesFromConfig,
    QuickActionError,
    type QuickCommand,
  } from '../features/quickactions';
  import type { SearchBarOptions } from '../features/search/usecase/ports';
  import { describeCommand } from './quickCapture';
  import { t } from '../shared/i18n';
  import Markdown from '../shared/ui/Markdown.svelte';
  import { errorMessage } from '../shared/api-client/http';
  import type { Container } from './container';
  import { itemTarget, type QuickItem, search } from './quickSearch';
  import { answerTarget, searchUsage } from './searchUsage';
  import { quickItemLoaders } from './quickItems';
  import { createSmartSearch } from './searchSources';
  import { buildRows, keyIntent, move, type Row } from './searchRows';
  import AnswerCard from '../features/search/ui/answers/AnswerCard.svelte';
  import type { ActionItem, Answer } from '../features/search/domain/smart/answer';
  import type { WeatherConfig } from '../features/weather/domain/weather';
  import { requestWidgetFocus } from '../shared/ui/widgetFocus';
  import { copyToClipboard } from '../shared/ui/clipboard';

  /**
   * Results of the hero search field (#222): a panel directly below the
   * field instead of a dialog. The field keeps the text and the focus; this
   * component gets the text, renders the groups of searchRows.ts and handles
   * the keys the field forwards. Loaded on first use (#116).
   */
  interface Props {
    container: Container;
    /** Enabled modules; disabled ones are not searched. */
    modules: Set<string>;
    /** Id of the listbox, referenced by the field's aria-controls. */
    listboxId: string;
    /** The text of the field. */
    query: string;
    /** Id of the selected option, for the field's aria-activedescendant. */
    activeId?: string;
    /** Whether results are shown, for the field's aria-expanded. */
    expanded?: boolean;
    /** Esc: close the panel, keep the text. */
    onclose: () => void;
    /** After opening a result elsewhere: close the panel and clear the field. */
    ondone: () => void;
    /** Location of the first weather widget for weather answers (#156). */
    weatherLocation?: WeatherConfig | null;
    /** First widget config of a kind, e.g. the transit stop (#161). */
    widgetConfig?: (kind: string) => Record<string, unknown> | null;
  }
  let {
    container,
    modules,
    listboxId,
    query = $bindable(),
    activeId = $bindable(),
    expanded = $bindable(),
    onclose,
    ondone,
    weatherLocation = null,
    widgetConfig = () => null,
  }: Props = $props();

  // svelte-ignore state_referenced_locally
  const smart = createSmartSearch(container, modules, {
    canAct: !container.auth.guest,
    weatherLocation: () => weatherLocation,
    widgetConfig: (kind) => widgetConfig(kind),
    score: (q, target, score) => searchUsage.score(q, target, score),
  });

  let items = $state<QuickItem[]>([]);
  let notes = new Map<string, { title: string; body: string }>();
  let todos = new Map<string, Parameters<Container['todos']['toggle']>[0]>();
  let active = $state(0);
  let status = $state('');
  let note = $state<{ title: string; body: string } | null>(null);
  let undoLast = $state<(() => Promise<void>) | null>(null);
  let openLast = $state<string | null>(null);

  // Options of the search field (#221) and the name of the search engine.
  let bar = $state<SearchBarOptions | null>(null);
  let engine = $state('');
  // svelte-ignore state_referenced_locally
  if (modules.has('search'))
    container.search
      .settings()
      .then((s) => {
        bar = s.bar;
        engine = s.providers.find((p) => p.id === s.defaultProvider)?.name ?? '';
      })
      .catch(() => {});
  const prefixes = $derived(mergePrefixes(palettePrefixes, prefixesFromConfig(bar?.capturePrefixes)));

  // "t Milch" → to-do; the same text as a note is the Ctrl+Enter alternative.
  const command = $derived(container.quickActions.parse(query, prefixes));
  const alternative = $derived.by((): QuickCommand | null => {
    const rest = captureRest(query, prefixes);
    if (!command || rest === null) return null;
    const other = command.kind === 'note' ? 'todo:' : 'note:';
    return container.quickActions.parse(`${other} ${rest}`, palettePrefixes);
  });

  // Body matches from the notes full-text search (#73).
  let noteHits = $state(new Map<string, string>());
  const searchable = $derived(
    items.map((i) => {
      const snippet = i.kind === 'note' ? noteHits.get(i.id) : undefined;
      return snippet === undefined ? i : { ...i, detail: snippet, keywords: [...(i.keywords ?? []), query] };
    }),
  );
  // Results chosen before for this query rank higher (#159).
  const groups = $derived(
    command ? [] : search(searchable, query, 5, (it, s) => searchUsage.score(query, itemTarget(it), s)),
  );
  const rowGroups = $derived(
    buildRows({ command, alternative, groups, web: modules.has('search') && query.trim() !== '' }),
  );
  const flat = $derived(rowGroups.flatMap((g) => g.rows));
  const optionId = (i: number) => `${listboxId}-option-${i}`;

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

  $effect(() => {
    // Load everything once when the panel opens; one failing source does not
    // block the others. The module calls live in quickItems.ts (#156).
    for (const load of quickItemLoaders(container, modules)) {
      load
        .then((r) => {
          items = [...items, ...r.items];
          if (r.notes) notes = r.notes;
          if (r.todos) todos = r.todos;
        })
        .catch(() => {});
    }
    void smart.prepare();
  });

  // Smart answers (#156): debounced, the newest query wins.
  let answers = $state<Answer[]>([]);
  $effect(() => {
    const q = query;
    if (command || q.trim().length < 2) {
      answers = [];
      return;
    }
    const ctrl = new AbortController();
    const timer = setTimeout(() => {
      smart
        .answer(q, ctrl.signal)
        .then((a) => (answers = a))
        .catch(() => {});
    }, 150);
    return () => {
      clearTimeout(timer);
      ctrl.abort();
    };
  });

  // Typing starts over: first row selected, the last confirmation goes away.
  $effect(() => {
    if (!query) return; // keep the confirmation after a command cleared the field
    active = 0;
    status = '';
    undoLast = null;
    openLast = null;
  });

  const visible = $derived(query.trim() !== '' || status !== '' || note !== null);
  $effect(() => {
    activeId = visible && flat.length ? optionId(Math.min(active, flat.length - 1)) : undefined;
    expanded = visible && (flat.length > 0 || answers.length > 0);
  });

  function searchWeb() {
    const target = modules.has('search') ? container.search.targetFor(query) : null;
    if (!target) return;
    if (bar?.openInNewTab) {
      window.open(target, '_blank', 'noopener');
      ondone();
    } else window.location.assign(target);
  }

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
      openLast = res.openUrl ?? null;
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
    openLast = null; // the bookmark is gone
    if (!u) return;
    try {
      await u();
      status = t('quick.undone');
    } catch (err) {
      status = errorMessage(err);
    }
  }

  async function choose(item: QuickItem) {
    // Only a chosen result is remembered, never the query alone (#159).
    searchUsage.record(query, itemTarget(item));
    if (item.url) {
      window.open(item.url, '_blank', 'noopener,noreferrer');
      ondone();
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

  function run(row: Row | undefined) {
    if (!row) return searchWeb();
    switch (row.kind) {
      case 'capture':
      case 'action':
        return void runCommand(row.command);
      case 'result':
        return void choose(row.item);
      case 'web':
        return searchWeb();
    }
  }

  async function execute(item: ActionItem, intent: string) {
    searchUsage.record(query, answerTarget(intent));
    try {
      await smart.execute(item);
      status = t('smart.action_done', { name: item.target.name });
      answers = await smart.answer(query, new AbortController().signal);
    } catch (err) {
      status = errorMessage(err);
    }
  }

  function openWidget(kind: string, intent: string) {
    searchUsage.record(query, answerTarget(intent));
    requestWidgetFocus([kind]);
    ondone();
  }

  async function copyResult(text: string) {
    status = (await copyToClipboard(text)) ? `${t('smart.copied')}: ${text}` : '';
  }

  /** Handles a key the field forwards; returns whether it was used. */
  export function keydown(e: KeyboardEvent): boolean {
    const intent = keyIntent(e);
    switch (intent) {
      case 'down':
      case 'up':
        active = move(active, flat.length, intent);
        document.getElementById(optionId(active))?.scrollIntoView({ block: 'nearest' });
        return true;
      case 'close':
        if (note) note = null;
        else onclose();
        return true;
      case 'web':
        searchWeb();
        return true;
      case 'alternative':
        if (alternative) void runCommand(alternative);
        return true;
      case 'run': {
        // A calculator result is copied (#160) while its row is the default.
        const calc =
          command || active > 0 ? undefined : smart.direct(query).find((a) => a.copy !== undefined);
        if (calc?.copy !== undefined) void copyResult(calc.copy);
        else run(flat[active]);
        return true;
      }
      default:
        return false;
    }
  }
</script>

{#if visible}
  <div class="panel">
    {#if note}
      <div class="note">
        <button class="btn btn-icon" type="button" onclick={() => (note = null)}>{t('app.back')}</button>
        <h4>{note.title}</h4>
        <Markdown source={note.body} />
      </div>
    {:else}
      {#if answers.length > 0}
        <!-- Instant answers above the rows (#156). -->
        <section class="answers" aria-label={t('smart.answers')} aria-live="polite">
          {#each answers as a (a.intent + a.title)}
            <AnswerCard
              answer={a}
              onopen={(kind) => openWidget(kind, a.intent)}
              onexecute={(item) => execute(item, a.intent)}
              onpick={(name) => (query = `${query.trim()} ${name}`)}
            />
          {/each}
        </section>
      {/if}
      {#if flat.length}
        <div id={listboxId} role="listbox" aria-label={t('app.results')} class="rows">
          {#each rowGroups as g (g.id)}
            <div role="group" aria-labelledby="{listboxId}-group-{g.id}">
              <div class="group" id="{listboxId}-group-{g.id}">{t(g.label)}</div>
              {#each g.rows as row (row.id)}
                {@const i = flat.indexOf(row)}
                <!-- The field keeps the focus; rows are reached with the arrow keys. -->
                <div
                  id={optionId(i)}
                  role="option"
                  tabindex="-1"
                  aria-selected={i === active}
                  class="row"
                  class:active={i === active}
                  onmousedown={(e) => e.preventDefault()}
                  onclick={() => run(row)}
                  onkeydown={(e) => e.key === 'Enter' && run(row)}
                  onmousemove={() => (active = i)}
                >
                  {#if row.kind === 'capture'}
                    <span class="title">{describeCommand(row.command, container.quickActions)}</span>
                    <kbd>{row.shortcut}</kbd>
                  {:else if row.kind === 'result'}
                    <span class="title">{row.item.title}</span>
                    {#if row.item.detail}<span class="muted detail">{row.item.detail}</span>{/if}
                  {:else if row.kind === 'action'}
                    <span class="title">{describeCommand(row.command, container.quickActions)}</span>
                    <span class="muted detail"
                      >{row.command.kind === 'restart' ? 'Docker' : 'Wake-on-LAN'}</span
                    >
                  {:else}
                    <span class="title"
                      >{engine
                        ? t('app.search_web_with', { query: query.trim(), engine })
                        : t('app.search_web', { query: query.trim() })}</span
                    >
                    <kbd>Alt Enter</kbd>
                  {/if}
                </div>
              {/each}
            </div>
          {/each}
        </div>
      {/if}
      <p class="muted hint" role="status">
        {status}
        {#if openLast}<a href={openLast} target="_blank" rel="noopener noreferrer">{t('quick.open')}</a>{/if}
        {#if undoLast}<button class="btn" type="button" onclick={undo}>{t('dashboard.undo')}</button>{/if}
      </p>
      <p class="sr-only" aria-live="polite">
        {query.trim() ? t('app.results_count', { n: flat.length }) : ''}
      </p>
    {/if}
  </div>
{/if}

<style>
  .panel {
    position: absolute;
    top: calc(100% + var(--space-2));
    left: 0;
    right: 0;
    z-index: 40;
    max-height: min(70vh, 34rem);
    overflow: auto;
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: 14px;
    background: var(--color-surface);
    box-shadow: 0 12px 36px color-mix(in srgb, #000 16%, transparent);
    text-align: left;
  }
  .answers {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .group {
    padding: var(--space-2) var(--space-2) 2px;
    font-size: 0.75em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2);
    border-radius: 8px;
    cursor: pointer;
  }
  .row.active {
    background: color-mix(in srgb, var(--color-accent) 14%, transparent);
  }
  .title {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .detail {
    flex: 0 1 auto;
    min-width: 0;
    max-width: 40%;
    font-size: 0.8em;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  kbd {
    flex-shrink: 0;
    font-size: 0.75rem;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    padding: 0 6px;
    color: var(--color-text-muted);
  }
  .hint {
    font-size: 0.8em;
    margin: var(--space-2) var(--space-2) 0;
  }
  .hint:empty {
    display: none;
  }
  .note h4 {
    margin: var(--space-2) 0;
  }
  @media (max-width: 600px) {
    /* No keyboard on a phone: the shortcut hints only take room. */
    kbd {
      display: none;
    }
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
</style>
