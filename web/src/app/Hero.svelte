<script lang="ts">
  import { dayPart, greetingKey } from '../features/appearance/domain/greeting';
  import { HINT_INTERVAL_MS, nextHint, SEARCH_HINTS } from '../features/search/domain/hints';
  import { t } from '../shared/i18n';
  import Icon from '../shared/ui/icons/Icon.svelte';

  /** Props of the hero below the top bar (#189): greeting, search pill and tabs. */
  interface Props {
    /** Name in the greeting; may be empty. */
    name: string;
    /** Called with the text typed so far; opens the quick search. */
    onsearch: (text: string) => void;
    /** Id of the element the dashboard renders its tab pills into. */
    tabsId: string;
  }
  let { name, onsearch, tabsId }: Props = $props();

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

  function take(input: HTMLInputElement) {
    const text = input.value;
    input.value = '';
    onsearch(text);
  }
</script>

<section class="hero">
  <h1>{hello}</h1>
  <label class="pill" data-tour="search">
    <Icon name="search" size={20} />
    <input
      type="search"
      aria-label={t('app.quick_search_ctrl_k_or')}
      aria-keyshortcuts="Control+K /"
      placeholder={t(SEARCH_HINTS[hint] ?? SEARCH_HINTS[0])}
      autocomplete="off"
      onfocus={() => (focused = true)}
      onblur={() => (focused = false)}
      oninput={(e) => take(e.currentTarget)}
      onkeydown={(e) => e.key === 'Enter' && take(e.currentTarget)}
    />
    <kbd>{t('app.ctrl_k')}</kbd>
  </label>
  <div id={tabsId} class="tabs"></div>
</section>

<style>
  .hero {
    display: grid;
    justify-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4) var(--space-4);
    max-width: 1600px;
    margin: 0 auto;
  }
  h1 {
    margin: 0;
    font-size: clamp(1.6rem, 3vw, 2.2rem);
    font-weight: 700;
    text-align: center;
  }
  .pill {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: min(40rem, 100%);
    height: 3.2rem;
    padding: 0 var(--space-4);
    border-radius: 999px;
    background: var(--color-surface);
    box-shadow: var(--shadow);
    color: var(--color-text-muted);
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
