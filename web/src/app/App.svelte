<script lang="ts">
  import { t } from '../shared/i18n';
  import { LoginPage, SetupPage, type AuthView } from '../features/auth';
  import { errorMessage } from '../shared/api-client/http';
  import { canStartOffline, clearOfflineData, rememberSignedIn } from '../shared/pwa';
  import { createContainer } from './container';
  import Shell from './Shell.svelte';

  let view = $state<AuthView | 'loading'>('loading');
  let loadError = $state('');

  const container = createContainer(() => (view = 'login'));

  async function refresh() {
    try {
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
  <SetupPage auth={container.auth} ondone={refresh} />
{:else if view === 'login'}
  <LoginPage auth={container.auth} ondone={refresh} />
{:else if view === 'app'}
  {#if container.auth.demo}
    <p class="demo-banner" role="note">{t('app.demo_banner')}</p>
  {/if}
  <Shell {container} onlogout={refresh} />
{/if}

<style>
  .demo-banner {
    margin: 0;
    padding: var(--space-1) var(--space-3);
    text-align: center;
    font-size: 0.9em;
    background: var(--color-accent);
    color: var(--color-on-accent, #fff);
  }
</style>
