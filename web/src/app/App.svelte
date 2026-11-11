<script lang="ts">
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
  <p class="error" role="alert">Could not reach the server: {loadError}</p>
{:else if view === 'setup'}
  <SetupPage auth={container.auth} ondone={refresh} />
{:else if view === 'login'}
  <LoginPage auth={container.auth} ondone={refresh} />
{:else if view === 'app'}
  <Shell {container} onlogout={refresh} />
{/if}
