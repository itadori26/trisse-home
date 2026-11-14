<script lang="ts">
  import { t } from '../shared/i18n';
  import Icon from '../shared/ui/icons/Icon.svelte';

  /** Props of the account menu of the top bar (#189): edit, settings, log out. */
  interface Props {
    /** Initial shown in the avatar. */
    initial: string;
    editing: boolean;
    /** Guests cannot edit the layout. */
    canEdit: boolean;
    onedit: () => void;
    onsettings: () => void;
    onlogout: () => void;
  }
  let { initial, editing, canEdit, onedit, onsettings, onlogout }: Props = $props();

  let open = $state(false);
  let trigger: HTMLButtonElement;
  let menu = $state<HTMLDivElement>();

  const items = () => [...(menu?.querySelectorAll<HTMLButtonElement>('[role="menuitem"]') ?? [])];

  function close(focusTrigger = true) {
    open = false;
    if (focusTrigger) trigger.focus();
  }

  function choose(action: () => void) {
    close(false);
    action();
  }

  // Focus the first item when the menu opens.
  $effect(() => {
    if (open) items()[0]?.focus();
  });

  function onkeydown(e: KeyboardEvent) {
    const list = items();
    const i = list.indexOf(document.activeElement as HTMLButtonElement);
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      close();
    } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      const next = (i + (e.key === 'ArrowDown' ? 1 : list.length - 1)) % list.length;
      list[next]?.focus();
    } else if (e.key === 'Home' || e.key === 'End') {
      e.preventDefault();
      list[e.key === 'Home' ? 0 : list.length - 1]?.focus();
    } else if (e.key === 'Tab') {
      close(false);
    }
  }

  // Clicks outside close the menu.
  $effect(() => {
    if (!open) return;
    const away = (e: PointerEvent) => {
      const target = e.target as Node;
      if (!menu?.contains(target) && !trigger.contains(target)) close(false);
    };
    document.addEventListener('pointerdown', away);
    return () => document.removeEventListener('pointerdown', away);
  });
</script>

<div class="account">
  <!-- The tour points at this button for "Edit mode" and "Settings". -->
  <span data-tour="settings"
    ><button
      bind:this={trigger}
      class="trigger"
      type="button"
      data-tour="edit"
      aria-haspopup="menu"
      aria-expanded={open}
      aria-controls="account-menu"
      aria-label={t('app.account')}
      title={t('app.account')}
      onclick={() => (open = !open)}
    >
      <span class="avatar" aria-hidden="true">{initial}</span>
      <Icon name="chevron" size={16} />
    </button></span
  >
  {#if open}
    <div
      id="account-menu"
      class="menu card"
      role="menu"
      aria-label={t('app.account')}
      bind:this={menu}
      tabindex="-1"
      {onkeydown}
    >
      {#if canEdit}
        <button type="button" role="menuitem" onclick={() => choose(onedit)}>
          <Icon name={editing ? 'done' : 'pencil'} size={18} />
          {editing ? t('app.done') : t('app.edit')}
        </button>
      {/if}
      <button type="button" role="menuitem" onclick={() => choose(onsettings)}>
        <Icon name="gear" size={18} />
        {t('app.settings')}
      </button>
      <button type="button" role="menuitem" onclick={() => choose(onlogout)}>
        <Icon name="power" size={18} />
        {t('app.log_out')}
      </button>
    </div>
  {/if}
</div>

<style>
  .account {
    position: relative;
  }
  .trigger {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px 3px 3px;
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: var(--color-surface);
    color: var(--color-text-muted);
    cursor: pointer;
  }
  .trigger:hover,
  .trigger[aria-expanded='true'] {
    border-color: var(--color-accent);
  }
  .avatar {
    width: 1.8rem;
    height: 1.8rem;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    font-weight: 600;
  }
  .menu {
    position: absolute;
    right: 0;
    top: calc(100% + var(--space-1));
    z-index: 40;
    min-width: 12rem;
    padding: var(--space-1);
    display: grid;
    box-shadow: 0 12px 40px rgb(0 0 0 / 18%);
  }
  [role='menuitem'] {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border: none;
    border-radius: var(--radius-small);
    background: none;
    color: var(--color-text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  [role='menuitem']:hover,
  [role='menuitem']:focus-visible {
    background: color-mix(in srgb, var(--color-accent) 10%, transparent);
    outline: none;
  }
</style>
