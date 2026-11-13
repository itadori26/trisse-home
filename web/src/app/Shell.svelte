<script lang="ts">
  import MaintenanceBanner from '../features/notifications/ui/MaintenanceBanner.svelte';
  import { setLanguage, t } from '../shared/i18n';
  import { type Appearance, cssVariables, defaultAppearance } from '../features/appearance';
  import { errorMessage } from '../shared/api-client/http';
  import { applyTheme } from '../shared/theme/theme';
  import type { Container } from './container';
  import { provideRefresh } from '../shared/refresh';
  import DashboardPage from './DashboardPage.svelte';
  import { attachShortcuts } from '../shared/shortcuts';
  import Icon from '../shared/ui/icons/Icon.svelte';
  import Greeting from './Greeting.svelte';
  import { firstWidgetConfig, type Tab } from '../features/dashboard';
  import { parseWeatherConfig } from '../features/weather/domain/weather';

  /** Props of the authenticated application shell. */
  interface Props {
    container: Container;
    /** Called after logout so the app can show the login screen. */
    onlogout: () => void;
  }
  let { container, onlogout }: Props = $props();
  // Components outside widgets (e.g. the maintenance banner) poll through it too.
  provideRefresh({
    get scheduler() {
      return container.refresh;
    },
    track: () => () => {},
  });

  let settingsOpen = $state(false);
  let editing = $state(false);
  let error = $state('');
  let appearance = $state<Appearance>(defaultAppearance);
  let modules = $state<Set<string> | null>(null);
  let widgetKinds = $state<Set<string>>(new Set());
  let touring = $state(false);
  let quickSearch = $state(false);
  let shortcutHelp = $state(false);
  let dashboard = $state<{ selectTabAt: (index: number) => void; undo: () => void; redo: () => void }>();
  let menuOpen = $state(false);
  let searchStart = $state('');
  let dashboardState = $state<{ tabs: Tab[]; activeId: string }>({ tabs: [], activeId: '' });

  // The greeting line shows the weather of the first weather widget (#167).
  const weatherLocation = $derived.by(() => {
    const cfg = firstWidgetConfig(dashboardState.tabs, dashboardState.activeId, 'weather');
    return cfg ? parseWeatherConfig(cfg) : null;
  });

  /** Typing into the bar's search field continues in the quick search (#167). */
  function openSearch(input: HTMLInputElement) {
    // Collect what is typed while the quick search is still loading.
    searchStart += input.value;
    input.value = '';
    quickSearch = true;
  }

  // Undo/redo shortcuts only exist in edit mode (#93).
  $effect(() => {
    if (!editing) return;
    const offs = [
      container.shortcuts.register({
        id: 'edit.undo',
        combos: ['Mod+Z'],
        display: [t('shortcuts.ctrl_z')],
        label: 'shortcuts.undo',
        run: () => dashboard?.undo(),
      }),
      container.shortcuts.register({
        id: 'edit.redo',
        combos: ['Mod+Shift+Z', 'Mod+Y'],
        display: [t('shortcuts.ctrl_shift_z'), t('shortcuts.ctrl_y')],
        label: 'shortcuts.redo',
        run: () => dashboard?.redo(),
      }),
    ];
    return () => offs.forEach((off) => off());
  });

  // Global shortcuts (#95): one registry, one listener.
  $effect(() => {
    const r = container.shortcuts;
    const offs = [
      r.register({
        id: 'search.open',
        combos: ['Mod+K', '/'],
        display: [t('app.ctrl_k'), '/'],
        label: 'shortcuts.search',
        typingCombos: ['Mod+K'],
        run: () => (quickSearch = true),
      }),
      r.register({
        id: 'tabs.select',
        combos: ['1', '2', '3', '4', '5', '6', '7', '8', '9'],
        display: ['1–9'],
        label: 'shortcuts.tabs',
        run: (combo) => dashboard?.selectTabAt(Number(combo) - 1),
      }),
      ...(container.auth.guest
        ? []
        : [
            r.register({
              id: 'edit.toggle',
              combos: ['e'],
              display: ['E'],
              label: 'shortcuts.edit',
              run: () => (editing = !editing),
            }),
          ]),
      r.register({
        id: 'shortcuts.help',
        combos: ['?'],
        display: ['?'],
        label: 'shortcuts.help',
        run: () => (shortcutHelp = true),
      }),
    ];
    const detach = attachShortcuts(r);
    return () => {
      detach();
      offs.forEach((off) => off());
    };
  });
  let offline = $state(!navigator.onLine);

  $effect(() => applyTheme(appearance.theme));
  $effect(() => setLanguage(appearance.language));

  $effect(() => {
    const root = document.documentElement;
    const vars = cssVariables(appearance);
    for (const [k, v] of Object.entries(vars)) root.style.setProperty(k, v);
    return () => Object.keys(vars).forEach((k) => root.style.removeProperty(k));
  });

  $effect(() => {
    // textContent (not innerHTML) so user CSS can never inject markup.
    const style = document.createElement('style');
    style.dataset['source'] = 'custom-css';
    style.textContent = appearance.customCss;
    document.head.append(style);
    return () => style.remove();
  });

  async function init() {
    try {
      const info = await container.enabledModules();
      modules = info.modules;
      widgetKinds = info.widgetKinds;
      if (info.modules.has('appearance')) appearance = await container.appearance.load();
      // Wait a moment so the dashboard is rendered before the tour measures it.
      if (
        !container.auth.guest &&
        info.modules.has('onboarding') &&
        (await container.onboarding.shouldShowTour())
      )
        setTimeout(startTour, 600);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function startTour() {
    editing = false;
    touring = true;
  }

  function finishTour() {
    touring = false;
    container.onboarding.finish().catch((err) => (error = errorMessage(err)));
  }

  async function restartTour() {
    settingsOpen = false;
    try {
      await container.onboarding.restart();
      setTimeout(startTour, 200);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function logout() {
    try {
      await container.auth.logout();
      onlogout();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  init();
</script>

<svelte:window ononline={() => (offline = false)} onoffline={() => (offline = true)} />

<!-- One compact top bar: brand, tabs, search and actions (#167). -->
<header class="topbar">
  <strong class="brand">trisse-home</strong>
  <div id="topbar-tabs" class="tabs-slot"></div>
  <label class="search">
    <Icon name="search" size={16} />
    <input
      type="search"
      aria-label={t('app.quick_search_ctrl_k_or')}
      aria-keyshortcuts="Control+K /"
      placeholder={t('app.search_placeholder')}
      autocomplete="off"
      oninput={(e) => openSearch(e.currentTarget)}
      onkeydown={(e) => e.key === 'Enter' && openSearch(e.currentTarget)}
    />
    <kbd>{t('app.ctrl_k')}</kbd>
  </label>
  <button
    class="btn btn-icon menu"
    type="button"
    aria-expanded={menuOpen}
    aria-controls="topbar-actions"
    aria-label={t('app.menu')}
    onclick={() => (menuOpen = !menuOpen)}><Icon name="menu" size={20} /></button
  >
  <nav id="topbar-actions" class="actions" class:open={menuOpen} aria-label={t('app.menu')}>
    {#if !container.auth.guest}
      <button
        class="btn btn-icon"
        class:on={editing}
        type="button"
        data-tour="edit"
        aria-pressed={editing}
        aria-label={editing ? t('app.done') : t('app.edit')}
        title={editing ? t('app.done') : t('app.edit')}
        onclick={() => ((editing = !editing), (menuOpen = false))}
        ><Icon name={editing ? 'done' : 'pencil'} size={20} /><span class="label"
          >{editing ? t('app.done') : t('app.edit')}</span
        ></button
      >
    {/if}
    <button
      class="btn btn-icon"
      type="button"
      data-tour="settings"
      aria-label={t('app.settings')}
      title={t('app.settings')}
      onclick={() => ((settingsOpen = true), (menuOpen = false))}
      ><Icon name="gear" size={20} /><span class="label">{t('app.settings')}</span></button
    >
    <button
      class="btn btn-icon"
      type="button"
      aria-label={t('app.log_out')}
      title={t('app.log_out')}
      onclick={logout}><Icon name="power" size={20} /><span class="label">{t('app.log_out')}</span></button
    >
  </nav>
</header>
{#if appearance.greeting}
  <Greeting
    name={appearance.greetingName}
    weather={modules?.has('weather') ? container.weather : null}
    location={weatherLocation}
  />
{/if}

<main class="content">
  {#if offline}<p class="offline" role="status">
      {t('app.you_are_offline_showing_the_last')}
    </p>{/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if modules}
    {#if modules.has('notifications') && !container.auth.guest}<MaintenanceBanner
        notifications={container.notifications}
      />{/if}
    <DashboardPage
      bind:this={dashboard}
      {container}
      {editing}
      {widgetKinds}
      tabBarTarget="topbar-tabs"
      onstate={(tabs, activeId) => (dashboardState = { tabs, activeId })}
    />
  {/if}
</main>

{#if settingsOpen && modules}
  <!-- Loaded on first use: the settings are not needed for the first paint (ADR 0011). -->
  {#await import('./SettingsDialog.svelte') then { default: SettingsDialog }}
    <SettingsDialog
      {container}
      {modules}
      {appearance}
      onappearance={(a) => (appearance = a)}
      onrestarttour={modules.has('onboarding') ? restartTour : undefined}
      onclose={() => (settingsOpen = false)}
    />
  {/await}
{/if}

{#if quickSearch && modules}
  {#await import('./QuickSearch.svelte') then { default: QuickSearch }}
    <QuickSearch
      {container}
      {modules}
      initialQuery={searchStart}
      {weatherLocation}
      widgetConfig={(kind) => firstWidgetConfig(dashboardState.tabs, dashboardState.activeId, kind)}
      onclose={() => ((quickSearch = false), (searchStart = ''))}
    />
  {/await}
{/if}

{#if shortcutHelp}
  {#await import('./ShortcutsDialog.svelte') then { default: ShortcutsDialog }}
    <ShortcutsDialog shortcuts={container.shortcuts.all()} onclose={() => (shortcutHelp = false)} />
  {/await}
{/if}

{#if touring}
  {#await import('../features/onboarding/ui/Tour.svelte') then { default: Tour }}
    <Tour onfinish={finishTour} />
  {/await}
{/if}

<style>
  .offline {
    margin: 0 0 var(--space-3);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-small);
    background: color-mix(in srgb, var(--color-warning) 22%, transparent);
  }
  /* One compact bar (#167): brand · tabs · search · actions. */
  .topbar {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) minmax(12rem, 22rem) auto;
    grid-template-areas: 'brand tabs search actions';
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-4);
    max-width: 1600px;
    margin: 0 auto;
  }
  .brand {
    grid-area: brand;
    letter-spacing: 0.02em;
    white-space: nowrap;
  }
  .tabs-slot {
    grid-area: tabs;
    min-width: 0;
  }
  .search {
    grid-area: search;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-2);
    height: 2.25rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    background: var(--color-surface);
    color: var(--color-text-muted);
  }
  .search:focus-within {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-accent) 30%, transparent);
  }
  .search input {
    flex: 1;
    min-width: 0;
    border: none;
    background: none;
    color: var(--color-text);
    font: inherit;
    outline: none;
  }
  kbd {
    font-size: 0.7em;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    padding: 0 4px;
  }
  .actions {
    grid-area: actions;
    display: flex;
    gap: var(--space-1);
  }
  .actions .label {
    display: none;
  }
  .menu {
    display: none;
  }
  .on {
    color: var(--color-accent);
  }
  /* Phones: brand, search and a menu button; the tabs scroll in a row below. */
  @media (max-width: 600px) {
    .topbar {
      grid-template-columns: auto minmax(0, 1fr) auto;
      grid-template-areas: 'brand search menu' 'tabs tabs tabs';
      gap: var(--space-2);
      padding: var(--space-2) var(--space-3);
      position: relative;
    }
    kbd {
      display: none;
    }
    .menu {
      display: inline-flex;
      grid-area: menu;
    }
    .actions {
      display: none;
      position: absolute;
      top: 100%;
      right: var(--space-3);
      z-index: 30;
      flex-direction: column;
      padding: var(--space-1);
      background: var(--color-surface);
      border-radius: var(--radius-small);
      box-shadow: var(--shadow);
    }
    .actions.open {
      display: flex;
    }
    .actions .btn {
      justify-content: flex-start;
      gap: var(--space-2);
    }
    .actions .label {
      display: inline;
    }
  }
  .content {
    padding: 0 var(--space-4) var(--space-6);
    max-width: 1600px;
    margin: 0 auto;
  }
</style>
