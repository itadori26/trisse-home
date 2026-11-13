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
  import AccountMenu from './AccountMenu.svelte';
  import Hero from './Hero.svelte';
  import TopInfo from './TopInfo.svelte';
  import { launcherSources } from './launcherSources';
  import { LauncherUseCases } from '../features/launcher/usecase/launcherUseCases';
  import { type Segment, statusSummary } from '../features/launcher/domain/launcher';
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
  let searchStart = $state('');
  let launcherOpen = $state<Segment | null>(null);
  let launcherTrigger = $state<HTMLButtonElement>();
  let serviceStatus = $state<ReturnType<typeof statusSummary>>({ kind: 'none', offline: 0 });
  const launcher = $derived(modules ? new LauncherUseCases(launcherSources(container, modules)) : null);

  // Status pill (#189): all online, or how many services are offline.
  $effect(() => {
    const l = launcher;
    if (!l?.hasServices) return;
    const poll = () =>
      l
        .services()
        .then((s) => (serviceStatus = statusSummary(s)))
        .catch(() => {});
    poll();
    const id = setInterval(poll, 60_000);
    return () => clearInterval(id);
  });

  function closeLauncher() {
    launcherOpen = null;
    launcherTrigger?.focus();
  }

  const accountInitial = $derived(appearance.greetingName.trim().charAt(0).toUpperCase());
  let dashboardState = $state<{ tabs: Tab[]; activeId: string }>({ tabs: [], activeId: '' });

  // The top bar shows the weather of the first weather widget (#167, #189).
  const weatherLocation = $derived.by(() => {
    const cfg = firstWidgetConfig(dashboardState.tabs, dashboardState.activeId, 'weather');
    return cfg ? parseWeatherConfig(cfg) : null;
  });

  /** Typing into the search pill continues in the quick search (#167, #189). */
  function openSearch(text: string) {
    // Collect what is typed while the quick search is still loading.
    searchStart += text;
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

<!-- Top bar and hero (#189): brand · date, weather, to-dos · status, launcher, account. -->
<header class="topbar">
  <a class="brand" href="/" aria-label="trisse-home"
    ><span class="logo" aria-hidden="true">t</span>trisse-home</a
  >
  <div class="center">
    <TopInfo
      weather={modules?.has('weather') ? container.weather : null}
      location={weatherLocation}
      todos={modules?.has('todos') ? container.todos : null}
    />
  </div>
  <div class="right">
    {#if serviceStatus.kind !== 'none'}
      <button
        type="button"
        class="status"
        data-kind={serviceStatus.kind}
        data-launcher-trigger
        onclick={() => (launcherOpen = launcherOpen ? null : 'services')}
      >
        <span class="dot" aria-hidden="true"></span>
        {serviceStatus.kind === 'ok'
          ? t('launcher.all_online')
          : t('launcher.some_offline', { n: serviceStatus.offline })}
      </button>
    {/if}
    {#if launcher}
      <button
        bind:this={launcherTrigger}
        type="button"
        class="round"
        data-launcher-trigger
        aria-haspopup="dialog"
        aria-expanded={launcherOpen !== null}
        aria-label={t('launcher.title')}
        title={t('launcher.title')}
        onclick={() => (launcherOpen = launcherOpen ? null : 'all')}><Icon name="grid" size={22} /></button
      >
    {/if}
    <AccountMenu
      initial={accountInitial}
      {editing}
      canEdit={!container.auth.guest}
      onedit={() => (editing = !editing)}
      onsettings={() => (settingsOpen = true)}
      onlogout={logout}
    />
  </div>
  {#if launcherOpen && launcher}
    {#await import('../features/launcher/ui/LauncherPanel.svelte') then { default: LauncherPanel }}
      <LauncherPanel
        {launcher}
        content={appearance.launcher}
        initialSegment={launcherOpen}
        onclose={closeLauncher}
        onsettings={() => ((launcherOpen = null), (settingsOpen = true))}
      />
    {/await}
  {/if}
</header>
<Hero name={appearance.greetingName} onsearch={openSearch} tabsId="header-tabs" />

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
      tabBarTarget="header-tabs"
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
  /* Top bar (#189): brand left, information centred, actions right. */
  .topbar {
    position: relative;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-4);
    max-width: 1600px;
    margin: 0 auto;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-weight: 700;
    color: inherit;
    text-decoration: none;
    white-space: nowrap;
  }
  .logo {
    width: 1.7rem;
    height: 1.7rem;
    border-radius: 6px;
    display: grid;
    place-items: center;
    background: var(--color-accent);
    color: #fff;
    font-size: 0.95rem;
  }
  .center {
    min-width: 0;
  }
  .right {
    justify-self: end;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 4px var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: var(--color-surface);
    color: var(--color-text);
    font: inherit;
    cursor: pointer;
    white-space: nowrap;
  }
  .status .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-success, #2e9d55);
  }
  .status[data-kind='offline'] {
    color: var(--color-danger);
  }
  .status[data-kind='offline'] .dot {
    background: var(--color-danger);
  }
  .round {
    width: 2.4rem;
    height: 2.4rem;
    border-radius: 50%;
    display: grid;
    place-items: center;
    border: 1px solid var(--color-border);
    background: var(--color-surface);
    color: var(--color-accent);
    cursor: pointer;
  }
  .round[aria-expanded='true'] {
    border-color: var(--color-accent);
  }
  /* Phones: brand and actions; the information line drops below. */
  @media (max-width: 900px) {
    .topbar {
      grid-template-columns: 1fr auto;
      padding: var(--space-2) var(--space-3);
    }
    .center {
      grid-column: 1 / -1;
      grid-row: 2;
      font-size: 0.85rem;
    }
  }
  @media (max-width: 600px) {
    .status {
      font-size: 0;
      padding: 8px;
    }
  }
  .content {
    padding: 0 var(--space-4) var(--space-6);
    max-width: 1600px;
    margin: 0 auto;
  }
</style>
