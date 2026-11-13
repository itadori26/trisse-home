<script lang="ts">
  import { t } from '../i18n';
  import { usesPlainHttp } from './plainHttp';

  /** Props of the plain-http warning below an address field (#191). */
  interface Props {
    /** The entered address; several addresses one per line. */
    url: string;
  }
  let { url }: Props = $props();
</script>

{#if usesPlainHttp(url)}
  <p class="warning" role="note">{t('common.plain_http_warning')}</p>
{/if}

<style>
  .warning {
    margin: calc(-1 * var(--space-1)) 0 var(--space-2);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-small);
    font-size: 0.85rem;
    background: color-mix(in srgb, var(--color-warning) 18%, transparent);
  }
</style>
