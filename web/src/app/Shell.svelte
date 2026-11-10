<script lang="ts">
  import { ChangePasswordForm } from '../features/auth';
  import { errorMessage } from '../shared/api-client/http';
  import { applyTheme, type ThemeMode } from '../shared/theme/theme';
  import Modal from '../shared/ui/Modal.svelte';
  import type { Container } from './container';

  /** Props of the authenticated application shell. */
  interface Props {
    container: Container;
    /** Called after logout so the app can show the login screen. */
    onlogout: () => void;
  }
  let { container, onlogout }: Props = $props();

  let settingsOpen = $state(false);
  let theme = $state<ThemeMode>('system');
  let error = $state('');

  $effect(() => applyTheme(theme));

  async function logout() {
    try {
      await container.auth.logout();
      onlogout();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<header class="topbar">
  <strong>trisse-home</strong>
  <nav>
    <button class="btn btn-icon" type="button" onclick={() => (settingsOpen = true)}>Settings</button>
    <button class="btn btn-icon" type="button" onclick={logout}>Log out</button>
  </nav>
</header>

<main class="content">
  {#if error}<p class="error">{error}</p>{/if}
  <section class="card placeholder" data-testid="dashboard">
    <h2>Dashboard</h2>
    <p class="muted">Widgets will appear here.</p>
  </section>
</main>

{#if settingsOpen}
  <Modal title="Settings" onclose={() => (settingsOpen = false)}>
    <h3>Theme</h3>
    <select class="input" bind:value={theme}>
      <option value="system">System</option>
      <option value="light">Light</option>
      <option value="dark">Dark</option>
    </select>
    <h3>Password</h3>
    <ChangePasswordForm auth={container.auth} />
  </Modal>
{/if}

<style>
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-4);
  }
  .content {
    padding: 0 var(--space-4) var(--space-6);
  }
  .placeholder {
    padding: var(--space-6);
  }
</style>
