<script lang="ts">
  import { dayPart, greetingKey } from '../features/appearance/domain/greeting';
  import { HINT_INTERVAL_MS, nextHint, SEARCH_HINTS } from '../features/search/domain/hints';
  import type { WeatherConfig } from '../features/weather/domain/weather';
  import { t } from '../shared/i18n';
  import Icon from '../shared/ui/icons/Icon.svelte';
  import type { Container } from './container';

  /** Props of the hero below the top bar (#189): greeting, search field and tabs. */
  interface Props {
    /** Name in the greeting; may be empty. */
    name: string;
    /** Id of the element the dashboard renders its tab pills into. */
    tabsId: string;
    container: Container;
    /** Enabled modules; null while loading (the search waits for them). */
    modules: Set<string> | null;
    /** Location of the first weather widget for weather answers (#156). */
    weatherLocation?: WeatherConfig | null;
    /** First widget config of a kind, e.g. the transit stop (#161). */
    widgetConfig?: (kind: string) => Record<string, unknown> | null;
  }
  let { name, tabsId, container, modules, weatherLocation = null, widgetConfig }: Props = $props();

  let now = $state(new Date());
  let hint = $state(0);
  let focused = $state(false);
  const reducedMotion =
    typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;

  $effect(() => {
    const id = setInterval(() => (now = new Date()), 60_000);
    return () => clearInterval(id);
  });

  // The placeholder rotates every few seconds; paused on focus, off for reduced motion.
  $effect(() => {
    if (reducedMotion) return;
    const id = setInterval(() => (hint = nextHint(hint, { focused, reducedMotion })), HINT_INTERVAL_MS);
    return () => clearInterval(id);
  });

  const hello = $derived(t(greetingKey(dayPart(now), name !== ''), { name }));

  // The field is the search (#222): it keeps the text and the focus, and the
  // results open in a panel directly below it (SearchPanel.svelte).
  const listboxId = 'hero-search-results';
  let query = $state('');
  let open = $state(false);
  let activeId = $state<string | undefined>();
  let expanded = $state(false);
  let input = $state<HTMLInputElement>();
  let box = $state<HTMLElement>();
  let panel = $state<{ keydown: (e: KeyboardEvent) => boolean }>();

  /** Focuses the field and opens the panel (Ctrl+K and /). */
  export function focus() {
    input?.focus();
    open = true;
  }

  function close() {
    open = false;
    expanded = false;
    activeId = undefined;
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      // First Esc closes the panel and keeps the text; the second clears and leaves.
      e.preventDefault();
      if (open) close();
      else {
        query = '';
        input?.blur();
      }
      return;
    }
    if (!open && e.key !== 'Tab') open = true;
    if (panel?.keydown(e)) e.preventDefault();
    else if (e.key === 'Enter' && !panel && modules?.has('search')) {
      // Enter before the panel has loaded still searches the web.
      const target = container.search.targetFor(query);
      if (target) window.location.assign(target);
    }
  }

  // A click outside the field and its panel closes the panel.
  $effect(() => {
    if (!open) return;
    const away = (e: PointerEvent) => {
      if (box && !box.contains(e.target as Node)) close();
    };
    document.addEventListener('pointerdown', away);
    return () => document.removeEventListener('pointerdown', away);
  });
</script>

<section class="hero">
  <h1>{hello}</h1>
  <div class="search" bind:this={box}>
    <label class="pill" class:active={open} data-tour="search">
      <Icon name="search" size={20} />
      <input
        bind:this={input}
        bind:value={query}
        type="search"
        role="combobox"
        aria-label={t('app.quick_search_ctrl_k_or')}
        aria-keyshortcuts="Control+K /"
        aria-expanded={expanded}
        aria-controls={listboxId}
        aria-autocomplete="list"
        aria-activedescendant={activeId}
        placeholder={t(SEARCH_HINTS[hint] ?? SEARCH_HINTS[0])}
        autocomplete="off"
        onfocus={() => ((focused = true), (open = true))}
        onblur={() => (focused = false)}
        oninput={() => (open = true)}
        {onkeydown}
      />
      <kbd>{open ? 'Esc' : t('app.ctrl_k')}</kbd>
    </label>
    {#if open && modules}
      <!-- Loaded on first use: the search logic is not part of the first paint (#116). -->
      {#await import('./SearchPanel.svelte') then { default: SearchPanel }}
        <SearchPanel
          bind:this={panel}
          bind:query
          bind:activeId
          bind:expanded
          {container}
          {modules}
          {listboxId}
          {weatherLocation}
          {widgetConfig}
          onclose={close}
          ondone={() => ((query = ''), close(), input?.blur())}
        />
      {/await}
    {/if}
  </div>
  <div id={tabsId} class="tabs"></div>
</section>

<style>
  .hero {
    display: grid;
    /* One column that may shrink below its content's width: otherwise the
       tab bar's min-content width stretches the column, and with it the
       greeting and the search, past the screen (#226). */
    grid-template-columns: minmax(0, 1fr);
    justify-items: center;
    /* Half the former gaps: search, tabs and widgets sit closer together. */
    gap: 6px;
    padding: var(--space-3) var(--space-4) var(--space-2);
    max-width: 1600px;
    margin: 0 auto;
  }
  h1 {
    margin: 0;
    font-size: clamp(1.6rem, 3vw, 2.2rem);
    font-weight: 700;
    text-align: center;
    overflow-wrap: anywhere; /* a long name wraps instead of widening the page */
  }
  .search {
    position: relative;
    width: min(40rem, 100%);
  }
  .pill {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    box-sizing: border-box;
    height: 3.2rem;
    padding: 0 var(--space-4);
    border: 2px solid transparent; /* the active border must not move the layout */
    border-radius: 999px;
    background: var(--color-surface);
    box-shadow: var(--shadow);
    color: var(--color-text-muted);
  }
  .pill.active {
    border-color: var(--color-accent);
  }
  .pill:focus-within {
    box-shadow:
      var(--shadow),
      0 0 0 2px color-mix(in srgb, var(--color-accent) 45%, transparent);
  }
  .pill input {
    flex: 1;
    min-width: 0;
    border: none;
    background: none;
    font: inherit;
    font-size: 1.05rem;
    color: var(--color-text);
    outline: none;
  }
  kbd {
    font-size: 0.75rem;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    padding: 0 6px;
    white-space: nowrap;
  }
  .tabs {
    max-width: 100%;
  }
  .tabs:empty {
    display: none;
  }
  @media (max-width: 600px) {
    .hero {
      padding: var(--space-2) var(--space-3) var(--space-3);
      gap: var(--space-2);
    }
    .pill {
      height: 2.8rem;
    }
    kbd {
      display: none;
    }
  }
</style>
