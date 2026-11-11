<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseTodosConfig } from '../domain/todo';

  /** Props of the to-do widget settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let showDone = $state(parseTodosConfig(config).showDone);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ showDone });
  }}
>
  <label><input type="checkbox" bind:checked={showDone} /> {t('todos.show_completed_to_dos')}</label>
  <p><button class="btn btn-primary" type="submit">{t('todos.save')}</button></p>
</form>
