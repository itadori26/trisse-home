<script lang="ts">
  import { type Appearance, cssVariables, defaultAppearance } from '../features/appearance';
  import { Tour } from '../features/onboarding';
  import { errorMessage } from '../shared/api-client/http';
  import { applyTheme } from '../shared/theme/theme';
  import type { Container } from './container';
  import DashboardPage from './DashboardPage.svelte';
  import QuickSearch from './QuickSearch.svelte';
  import { opensQuickSearch } from './quickSearch';
  import SettingsDialog from './SettingsDialog.svelte';

  /** Props of the authenticated application shell. */
  interface Props {
    container: Container;
    /** Called after logout so the app can show the login screen. */
    onlogout: () => void;
  }
  let { container, onlogout }: Props = $props();

  let settingsOpen = $state(false);
  let editing = $state(false);
  let error = $state('');
  let appearance = $state<Appearance>(defaultAppearance);
  let modules = $state<Set<string> | null>(null);
  let widgetKinds = $state<Set<string>>(new Set());
  let touring = $state(false);
  let quickSearch = $state(false);

  $effect(() => applyTheme(appearance.theme));

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
      if (info.modules.has('onboarding') && (await container.onboarding.shouldShowTour()))
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

<svelte:window
  onkeydown={(e) => {
    if (!quickSearch && !settingsOpen && opensQuickSearch(e, e.target)) {
      e.preventDefault();
      quickSearch = true;
    }
  }}
/>

<header class="topbar">
  <strong class="brand">trisse-home</strong>
  <nav>
    <button
      class="btn btn-icon"
      type="button"
      aria-keyshortcuts="Control+K /"
      title="Quick search (Ctrl+K or /)"
      onclick={() => (quickSearch = true)}>Search <kbd>Ctrl K</kbd></button
    >
    <button
      class="btn btn-icon"
      class:on={editing}
      type="button"
      data-tour="edit"
      aria-pressed={editing}
      onclick={() => (editing = !editing)}
    >
      {editing ? 'Done' : 'Edit'}
    </button>
    <button class="btn btn-icon" type="button" data-tour="settings" onclick={() => (settingsOpen = true)}
      >Settings</button
    >
    <button class="btn btn-icon" type="button" onclick={logout}>Log out</button>
  </nav>
</header>

<main class="content">
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if modules}
    <DashboardPage {container} {editing} {widgetKinds} />
  {/if}
</main>

{#if settingsOpen && modules}
  <SettingsDialog
    {container}
    {modules}
    {appearance}
    onappearance={(a) => (appearance = a)}
    onrestarttour={modules.has('onboarding') ? restartTour : undefined}
    onclose={() => (settingsOpen = false)}
  />
{/if}

{#if quickSearch && modules}
  <QuickSearch {container} {modules} onclose={() => (quickSearch = false)} />
{/if}

{#if touring}
  <Tour onfinish={finishTour} />
{/if}

<style>
  kbd {
    font-size: 0.7em;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    padding: 0 4px;
    margin-left: 4px;
  }
  @media (max-width: 600px) {
    kbd {
      display: none;
    }
  }
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-4);
    max-width: 1600px;
    margin: 0 auto;
  }
  .brand {
    letter-spacing: 0.02em;
  }
  .on {
    color: var(--color-accent);
    font-weight: 600;
  }
  .content {
    padding: 0 var(--space-4) var(--space-6);
    max-width: 1600px;
    margin: 0 auto;
  }
</style>
