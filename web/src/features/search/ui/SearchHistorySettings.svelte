<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { SearchUsage } from '../usecase/searchUsage';

  /** Props of the search history settings (#159); the history stays in this browser. */
  interface Props {
    usage: SearchUsage;
  }
  let { usage }: Props = $props();

  // svelte-ignore state_referenced_locally
  let enabled = $state(usage.enabled);
  // svelte-ignore state_referenced_locally
  let size = $state(usage.size);

  function toggle() {
    usage.setEnabled(enabled);
    size = usage.size;
  }

  function clear() {
    usage.clear();
    size = 0;
  }
</script>

<h3>{t('search.history_heading')}</h3>
<label class="check">
  <input type="checkbox" bind:checked={enabled} onchange={toggle} />
  {t('search.history_learn')}
</label>
<p class="muted">{t('search.history_hint')}</p>
<p>
  <button class="btn" type="button" onclick={clear} disabled={size === 0}>{t('search.history_clear')}</button>
  <span class="muted" role="status">{t('search.history_count', { n: size })}</span>
</p>

<style>
  .check {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  p {
    margin: var(--space-2) 0;
  }
</style>
