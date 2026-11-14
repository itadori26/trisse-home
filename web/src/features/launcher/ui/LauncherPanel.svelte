<script lang="ts">
  import { t } from '../../../shared/i18n';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { buildView, initial, type LauncherContent, safeUrl, type Segment } from '../domain/launcher';
  import type { LauncherData, LauncherUseCases } from '../usecase/launcherUseCases';

  /** Props of the launcher panel of the top bar (#189). */
  interface Props {
    launcher: LauncherUseCases;
    content: LauncherContent;
    /** Segment to open with, e.g. "services" from the status pill. */
    initialSegment?: Segment;
    onclose: () => void;
    /** Opens Settings → Appearance (launcher setting). */
    onsettings: () => void;
  }
  let { launcher, content, initialSegment = 'all', onclose, onsettings }: Props = $props();

  let data = $state<LauncherData>({ services: [], bookmarks: [] });
  let loaded = $state(false);
  let filter = $state('');
  // svelte-ignore state_referenced_locally
  let segment = $state<Segment>(initialSegment);
  let expanded = $state(new Set<'services' | 'bookmarks'>());
  let panel: HTMLDivElement;

  $effect(() => {
    launcher.load().then((d) => ((data = d), (loaded = true)));
  });

  const view = $derived(buildView(data.services, data.bookmarks, { filter, segment, content, expanded }));
  const segments = $derived(
    (['all', 'services', 'bookmarks'] as const).filter(
      (s) => s === 'all' || content === 'both' || content === s,
    ),
  );
  const segmentLabel: Record<Segment, 'launcher.all' | 'launcher.services' | 'launcher.bookmarks'> = {
    all: 'launcher.all',
    services: 'launcher.services',
    bookmarks: 'launcher.bookmarks',
  };

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      onclose();
    }
  }

  // Clicks outside the panel close it (the trigger toggles it itself).
  $effect(() => {
    const away = (e: PointerEvent) => {
      const target = e.target as Node | null;
      if (target && !panel.contains(target) && !(target as Element).closest?.('[data-launcher-trigger]'))
        onclose();
    };
    document.addEventListener('pointerdown', away);
    return () => document.removeEventListener('pointerdown', away);
  });
</script>

<div
  class="panel card"
  role="dialog"
  tabindex="-1"
  aria-label={t('launcher.title')}
  bind:this={panel}
  {onkeydown}
>
  <label class="filter">
    <Icon name="search" size={16} />
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="search"
      bind:value={filter}
      placeholder={t('launcher.filter')}
      aria-label={t('launcher.filter')}
      autofocus
    />
  </label>

  {#if content === 'both'}
    <div class="segments" role="group" aria-label={t('launcher.show')}>
      {#each segments as s (s)}
        <button type="button" class="chip" aria-pressed={segment === s} onclick={() => (segment = s)}>
          {t(segmentLabel[s])} <span class="count">{view.counts[s]}</span>
        </button>
      {/each}
    </div>
  {/if}

  {#if view.services}
    <section aria-labelledby="launcher-services">
      <header>
        <h3 id="launcher-services">{t('launcher.services')}</h3>
        {#if view.services.more}
          <button type="button" class="more" onclick={() => (expanded = new Set([...expanded, 'services']))}
            >{t('launcher.show_all_services', { n: view.services.total })}</button
          >
        {/if}
      </header>
      {#if view.services.items.length === 0}
        <p class="muted">{loaded ? t('launcher.no_services') : t('common.loading')}</p>
      {:else}
        <ul class="tiles">
          {#each view.services.items as s (s.id)}
            {@const href = safeUrl(s.url)}
            <li>
              <a class="tile" href={href ?? undefined} target="_blank" rel="noopener noreferrer">
                <strong>{s.name}</strong>
                <span class="status" data-state={s.state}>
                  <span class="dot" aria-hidden="true"></span>
                  {#if s.state === 'online'}{t('launcher.online_ms', {
                      ms: s.latencyMs,
                    })}{:else if s.state === 'offline'}{t('launcher.offline')}{:else}{t(
                      'launcher.unknown',
                    )}{/if}
                </span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  {#if view.bookmarks}
    <section aria-labelledby="launcher-bookmarks">
      <header>
        <h3 id="launcher-bookmarks">{t('launcher.bookmarks')}</h3>
        {#if view.bookmarks.more}
          <button type="button" class="more" onclick={() => (expanded = new Set([...expanded, 'bookmarks']))}
            >{t('launcher.show_all_bookmarks', { n: view.bookmarks.total })}</button
          >
        {/if}
      </header>
      {#if view.bookmarks.items.length === 0}
        <p class="muted">{loaded ? t('launcher.no_bookmarks') : t('common.loading')}</p>
      {:else}
        <ul class="links">
          {#each view.bookmarks.items as b (b.id)}
            {@const href = safeUrl(b.url)}
            {#if href}
              <li>
                <a class="link" {href} target="_blank" rel="noopener noreferrer">
                  <span class="avatar" aria-hidden="true">{initial(b.title)}</span>
                  <span class="title">{b.title}</span>
                  {#if b.folder}<span class="folder">{b.folder}</span>{/if}
                </a>
              </li>
            {/if}
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  <footer>
    <span
      >{t('launcher.footer_before')}
      <button type="button" class="settings-link" onclick={onsettings}>{t('launcher.footer_link')}</button
      ></span
    >
    <span class="muted">{t('launcher.footer_ctrl_k')}</span>
  </footer>
</div>

<style>
  .panel {
    position: absolute;
    top: calc(100% + var(--space-1));
    right: var(--space-4);
    z-index: 40;
    width: min(46rem, calc(100vw - 2 * var(--space-3)));
    max-height: min(80vh, 44rem);
    overflow: auto;
    padding: var(--space-3);
    display: grid;
    gap: var(--space-3);
    box-shadow: 0 12px 40px rgb(0 0 0 / 18%);
  }
  .filter {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-3);
    height: 2.6rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    color: var(--color-text-muted);
  }
  .filter:focus-within {
    border-color: var(--color-accent);
  }
  .filter input {
    flex: 1;
    min-width: 0;
    border: none;
    background: none;
    font: inherit;
    color: var(--color-text);
    outline: none;
  }
  .segments {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .chip {
    border: 1px solid var(--color-border);
    background: none;
    color: var(--color-text);
    border-radius: 999px;
    padding: 2px var(--space-3);
    font: inherit;
    cursor: pointer;
  }
  .chip[aria-pressed='true'] {
    background: var(--color-accent);
    border-color: var(--color-accent);
    color: var(--color-accent-contrast);
  }
  .count {
    opacity: 0.8;
  }
  section header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: var(--space-2);
  }
  h3 {
    margin: 0;
    font-size: 0.8rem;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--color-text-muted);
  }
  /* "Show all" is a small, muted text link, not a button (#189). */
  .more {
    border: none;
    background: none;
    padding: 0;
    font-size: 0.8rem;
    color: var(--color-text-muted);
    cursor: pointer;
  }
  .more:hover {
    text-decoration: underline;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
    gap: var(--space-2);
  }
  .tile {
    display: grid;
    gap: 2px;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    color: inherit;
    text-decoration: none;
  }
  .tile:hover,
  .tile:focus-visible {
    border-color: var(--color-accent);
  }
  .status {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    font-size: 0.85rem;
    color: var(--color-text-muted);
  }
  .status[data-state='offline'] {
    color: var(--color-danger);
    font-weight: 500;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-text-muted);
  }
  [data-state='online'] .dot {
    background: var(--color-success, #2e9d55);
  }
  [data-state='offline'] .dot {
    background: var(--color-danger);
  }
  .links {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(13rem, 1fr));
    gap: var(--space-1) var(--space-3);
  }
  .link {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1);
    border-radius: var(--radius-small);
    color: inherit;
    text-decoration: none;
  }
  .link:hover,
  .link:focus-visible {
    background: color-mix(in srgb, var(--color-accent) 8%, transparent);
  }
  .avatar {
    flex: none;
    width: 1.6rem;
    height: 1.6rem;
    border-radius: 50%;
    display: grid;
    place-items: center;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-accent);
    background: color-mix(in srgb, var(--color-accent) 12%, transparent);
  }
  .title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .folder {
    font-size: 0.8rem;
    color: var(--color-text-muted);
  }
  footer {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: var(--space-2);
    padding-top: var(--space-2);
    border-top: 1px solid var(--color-border);
    font-size: 0.85rem;
  }
  .settings-link {
    border: none;
    background: none;
    padding: 0;
    font: inherit;
    color: var(--color-accent);
    text-decoration: underline;
    cursor: pointer;
  }
  @media (max-width: 600px) {
    .panel {
      right: var(--space-2);
      left: var(--space-2);
      width: auto;
    }
  }
</style>
