<script lang="ts">
  import { loadLanguage, t } from '../shared/i18n';
  import type { AuthView } from '../features/auth';
  import { errorMessage } from '../shared/api-client/http';
  import { canStartOffline, clearOfflineData, rememberSignedIn } from '../shared/pwa';
  import { createContainer } from './container';
  // The dashboard shell is fetched right away, in parallel with the session
  // check, but kept out of the initial bundle (#116); sign-in pages load
  // only when needed.
  const shell = import('./Shell.svelte');

  let view = $state<AuthView | 'loading'>('loading');
  let loadError = $state('');

  const container = createContainer(() => (view = 'login'));

  async function refresh() {
    try {
      // The browser language's catalogue first, so the page never flashes English.
      await loadLanguage().catch(() => undefined);
      // A kiosk link (#50): the token is in the fragment, so it never reaches
      // the server log; it signs this browser in as the display.
      if (location.pathname === '/kiosk' && location.hash.length > 1) {
        const token = decodeURIComponent(location.hash.slice(1));
        history.replaceState(null, '', '/');
        await container.auth.kioskLogin(token);
      }
      view = await container.auth.currentView();
      loadError = '';
      rememberSignedIn(view === 'app');
      if (view !== 'app') await clearOfflineData();
    } catch (err) {
      // Offline after an earlier session: show the last loaded data read-only.
      if (canStartOffline()) view = 'app';
      else loadError = errorMessage(err);
    }
  }

  refresh();
</script>

{#if loadError}
  <p class="error" role="alert">{t('app.could_not_reach_the_server', { loadError: loadError })}</p>
{:else if view === 'setup'}
  {#await import('../features/auth/ui/SetupPage.svelte') then { default: SetupPage }}
    <SetupPage auth={container.auth} ondone={refresh} />
  {/await}
{:else if view === 'login'}
  {#await import('../features/auth/ui/LoginPage.svelte') then { default: LoginPage }}
    <LoginPage auth={container.auth} ondone={refresh} />
  {/await}
{:else if view === 'app' && location.pathname === '/capture'}
  <!-- Quick capture (#66): a page of its own, without the dashboard. -->
  {#await import('./CapturePage.svelte') then { default: CapturePage }}
    <CapturePage {container} />
  {/await}
{:else if view === 'app'}
  {#if container.auth.demo}
    <p class="demo-banner" role="note">{t('app.demo_banner')}</p>
  {/if}
  {#await shell then { default: Shell }}
    <Shell {container} onlogout={refresh} />
  {:catch err}
    <p class="error" role="alert">{t('app.could_not_reach_the_server', { loadError: errorMessage(err) })}</p>
  {/await}
{/if}

<style>
  .demo-banner {
    margin: 0;
    padding: var(--space-1) var(--space-3);
    text-align: center;
    font-size: 0.9em;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
  }
</style>
