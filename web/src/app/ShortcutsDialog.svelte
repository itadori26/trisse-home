<script lang="ts">
  import { type MessageKey, t } from '../shared/i18n';
  import Modal from '../shared/ui/Modal.svelte';
  import type { Shortcut } from '../shared/shortcuts';

  /** Props of the keyboard shortcut overview (#95). */
  interface Props {
    shortcuts: readonly Shortcut[];
    onclose: () => void;
  }
  let { shortcuts, onclose }: Props = $props();
</script>

<Modal title={t('shortcuts.title')} {onclose}>
  <table>
    <tbody>
      {#each shortcuts as s (s.id)}
        <tr>
          <td>
            {#each s.display as k, i (k)}{#if i > 0}<span class="or"> {t('shortcuts.or')} </span>{/if}<kbd
                >{k}</kbd
              >{/each}
          </td>
          <td>{t(s.label as MessageKey)}</td>
        </tr>
      {/each}
    </tbody>
  </table>
  <p class="muted">{t('shortcuts.hint')}</p>
</Modal>

<style>
  table {
    width: 100%;
    border-collapse: collapse;
  }
  td {
    padding: var(--space-1) var(--space-2);
    border-bottom: 1px solid var(--color-border);
  }
  td:first-child {
    white-space: nowrap;
    width: 1%;
  }
  .or {
    color: var(--color-text-muted);
    font-size: 0.85em;
  }
</style>
