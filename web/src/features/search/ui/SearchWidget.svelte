<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { QuickCapture } from '../usecase/quickCapture';
  import type { Answer } from '../domain/smart/answer';
  import AnswerCard from './answers/AnswerCard.svelte';
  import { requestWidgetFocus } from '../../../shared/ui/widgetFocus';
  import { copyToClipboard } from '../../../shared/ui/clipboard';
  import type { SearchUseCases } from '../usecase/searchUseCases';

  /** Props of the search bar widget. */
  interface Props {
    search: SearchUseCases;
    config: Record<string, unknown>;
    /** Quick capture with prefixes like "t " (#105); optional. */
    quick?: QuickCapture;
    /** Instant answers of the smart search (#156); optional. */
    answers?: (query: string, signal: AbortSignal) => Promise<Answer[]>;
  }
  let { search, config, quick, answers }: Props = $props();

  // Same engine as the quick search; debounced, the newest query wins.
  let found = $state<Answer[]>([]);
  let foundFor = '';
  $effect(() => {
    const q = query;
    if (!answers || preview || q.trim().length < 2) {
      found = [];
      return;
    }
    const ctrl = new AbortController();
    const timer = setTimeout(() => {
      answers(q, ctrl.signal)
        .then((a) => {
          found = a;
          foundFor = q;
        })
        .catch(() => {});
    }, 150);
    return () => {
      clearTimeout(timer);
      ctrl.abort();
    };
  });

  let query = $state('');
  const newTab = $derived(config['newTab'] === true);
  const preview = $derived(quick?.preview(query) ?? null);
  let done = $state<{ message: string; undo?: () => Promise<void>; openUrl?: string } | null>(null);
  let error = $state('');
  let doneTimer: ReturnType<typeof setTimeout> | undefined;

  function searchWeb() {
    const target = search.targetFor(query);
    if (!target) return;
    if (newTab) window.open(target, '_blank', 'noopener');
    else window.location.assign(target);
    query = '';
  }

  async function capture() {
    if (!quick) return;
    error = '';
    try {
      const res = await quick.run(query);
      if (!res) return searchWeb();
      query = '';
      done = res;
      clearTimeout(doneTimer);
      doneTimer = setTimeout(() => (done = null), 8000);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function submit(e: SubmitEvent) {
    e.preventDefault();
    // Enter copies a calculator result (#160); Shift+Enter still searches.
    const calc = foundFor === query ? found.find((a) => a.copy !== undefined) : undefined;
    if (preview) void capture();
    else if (calc?.copy !== undefined) void copyResult(calc.copy);
    else searchWeb();
  }

  async function copyResult(text: string) {
    if (!(await copyToClipboard(text))) return;
    done = { message: `${t('smart.copied')}: ${text}` };
    clearTimeout(doneTimer);
    doneTimer = setTimeout(() => (done = null), 3000);
  }

  function onkeydown(e: KeyboardEvent) {
    // Shift+Enter always searches the web, even with a quick prefix.
    if (e.key === 'Enter' && e.shiftKey) {
      e.preventDefault();
      searchWeb();
    }
  }

  async function undo() {
    const u = done?.undo;
    done = null;
    if (!u) return;
    try {
      await u();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form class="search" role="search" onsubmit={submit} data-tour="search">
  <input
    class="input"
    type="search"
    name="q"
    bind:value={query}
    {onkeydown}
    placeholder={t('search.search_bangs_like_w_work')}
    aria-label={t('search.search')}
    aria-describedby={preview ? 'search-quick-preview' : undefined}
    autocomplete="off"
  />
  {#if preview}
    <p id="search-quick-preview" class="hint">
      {t('quick.enter_will', { action: preview })} ·
      <span class="muted">{t('quick.shift_enter_searches')}</span>
    </p>
  {:else if done}
    <p class="hint done" role="status">
      {done.message}
      {#if done.openUrl}<a href={done.openUrl} target="_blank" rel="noopener noreferrer">{t('quick.open')}</a
        >{/if}
      {#if done.undo}<button class="link" type="button" onclick={undo}>{t('dashboard.undo')}</button>{/if}
    </p>
  {:else if error}
    <p class="hint error" role="alert">{error}</p>
  {/if}
  {#if found.length > 0}
    <section class="answers" aria-label={t('smart.answers')} aria-live="polite">
      {#each found as a (a.intent + a.title)}
        <AnswerCard answer={a} onopen={(kind) => requestWidgetFocus([kind])} />
      {/each}
    </section>
  {/if}
</form>

<style>
  /* Instant answers drop down below the bar without moving the layout (#156). */
  .answers {
    position: absolute;
    top: calc(100% - var(--space-2));
    left: var(--space-4);
    right: var(--space-4);
    z-index: 15;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
    gap: var(--space-2);
    padding: var(--space-2);
    border-radius: var(--radius);
    background: var(--color-bg, var(--color-surface));
    box-shadow: var(--shadow);
  }
  .search {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    position: relative;
  }
  .input {
    font-size: 1.15rem;
    padding: var(--space-3) var(--space-4);
    border-radius: 999px;
    background: var(--color-surface);
    box-shadow: var(--shadow);
  }
  .hint {
    position: absolute;
    left: var(--space-4);
    right: var(--space-4);
    bottom: -1.4em;
    margin: 0;
    font-size: 0.8rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .done {
    color: var(--color-success);
  }
  .link {
    border: 0;
    background: none;
    color: var(--color-accent);
    cursor: pointer;
    padding: 0 0 0 var(--space-1);
    font: inherit;
    text-decoration: underline;
  }
  .hint a {
    margin-left: var(--space-1);
  }
</style>
