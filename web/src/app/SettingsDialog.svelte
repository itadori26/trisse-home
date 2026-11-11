<script lang="ts">
  import { AdGuardSettings } from '../features/adguard';
  import { type Appearance, AppearanceSettings } from '../features/appearance';
  import { ChangePasswordForm } from '../features/auth';
  import { FoldersManager } from '../features/bookmarks';
  import { CalendarSettings } from '../features/calendar';
  import { FeedsManager } from '../features/rss';
  import { SearchSettings } from '../features/search';
  import { ServicesManager } from '../features/services';
  import Modal from '../shared/ui/Modal.svelte';
  import type { Container } from './container';

  /** Props of the global settings dialog. */
  interface Props {
    container: Container;
    modules: Set<string>;
    appearance: Appearance;
    onappearance: (a: Appearance) => void;
    /** Restarts the guided tour; omitted when the onboarding module is disabled. */
    onrestarttour?: () => void;
    onclose: () => void;
  }
  let { container, modules, appearance, onappearance, onrestarttour, onclose }: Props = $props();

  type Section = { id: string; label: string; module?: string };
  const sections: Section[] = [
    { id: 'appearance', label: 'Appearance', module: 'appearance' },
    { id: 'search', label: 'Search', module: 'search' },
    { id: 'services', label: 'Services', module: 'services' },
    { id: 'feeds', label: 'Feeds', module: 'rss' },
    { id: 'folders', label: 'Bookmark folders', module: 'bookmarks' },
    { id: 'calendar', label: 'Calendar', module: 'calendar' },
    { id: 'adguard', label: 'AdGuard', module: 'adguard' },
    { id: 'account', label: 'Account' },
  ];
  const visible = $derived(sections.filter((s) => !s.module || modules.has(s.module)));
  let active = $state('');
  const current = $derived(active || visible[0]?.id || 'account');
</script>

<Modal title="Settings" {onclose}>
  <nav class="sections">
    {#each visible as s (s.id)}
      <button class="btn" class:btn-primary={current === s.id} type="button" onclick={() => (active = s.id)}
        >{s.label}</button
      >
    {/each}
  </nav>
  {#if current === 'appearance'}
    <AppearanceSettings appearance={container.appearance} current={appearance} onchange={onappearance} />
  {:else if current === 'search'}
    <SearchSettings search={container.search} />
  {:else if current === 'services'}
    <ServicesManager services={container.services} />
  {:else if current === 'feeds'}
    <FeedsManager rss={container.rss} />
  {:else if current === 'folders'}
    <FoldersManager bookmarks={container.bookmarks} />
  {:else if current === 'calendar'}
    <CalendarSettings calendar={container.calendar} />
  {:else if current === 'adguard'}
    <AdGuardSettings adguard={container.adguard} />
  {:else}
    {#if onrestarttour}
      <h3>Guided tour</h3>
      <p class="muted">Shows the most important features of trisse-home step by step.</p>
      <button class="btn" type="button" onclick={onrestarttour}>Restart the tour</button>
    {/if}
    <h3>Change password</h3>
    <ChangePasswordForm auth={container.auth} />
  {/if}
</Modal>

<style>
  .sections {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin-bottom: var(--space-4);
  }
</style>
