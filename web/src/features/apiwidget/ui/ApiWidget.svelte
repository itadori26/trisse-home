<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatValue, parseWidgetConfig, type Values } from '../domain/apiwidget';
  import type { ApiWidgetUseCases } from '../usecase/apiWidgetUseCases';

  /** Props of the JSON widget (#48). */
  interface Props {
    apiwidget: ApiWidgetUseCases;
    config: Record<string, unknown>;
  }
  let { apiwidget, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseWidgetConfig(config));

  // The server requests the endpoint at most once per its interval; asking
  // it every minute only reads its cache.
  const POLL_MS = 60_000;
  let data = $state<Values | null>(null);
  let error = $state('');
  let gone = $state(false);

  $effect(() => {
    const id = cfg.endpointId;
    data = null;
    error = '';
    gone = false;
    if (!id) return;
    return poll(
      () =>
        apiwidget
          .values(id)
          .then((v) => ((data = v), (error = '')))
          .catch((err) => {
            gone = err instanceof ApiError && err.status === 404;
            error = errorMessage(err);
            throw err;
          }),
      POLL_MS,
    );
  });
</script>

{#if !cfg.endpointId}
  <p class="muted">{t('apiwidget.choose_in_settings')}</p>
{:else if gone}
  <p class="muted">{t('apiwidget.gone')}</p>
{:else if error && !data}
  <p class="error">{error}</p>
{:else if data}
  <p class="name">{data.name}</p>
  {#if data.error}<p class="error small" role="alert">{data.error}</p>{/if}
  <dl>
    {#each data.values as v, i (i)}
      <div class="row {v.status}" title={v.error || undefined}>
        <dt>{v.label}</dt>
        <dd>
          {#if v.error}<span class="muted">–</span>{:else}{formatValue(v, intlLocale())}{/if}
          {#if v.status === 'warning' || v.status === 'critical'}<span class="sr-only"
              >({t(v.status === 'warning' ? 'apiwidget.status_warning' : 'apiwidget.status_critical')})</span
            >{/if}
        </dd>
      </div>
    {/each}
  </dl>
{/if}

<style>
  .name {
    margin: 0 0 var(--space-1);
    font-size: 0.8em;
    color: var(--color-text-muted);
  }
  dl {
    margin: 0;
    display: grid;
    gap: 2px;
  }
  .row {
    display: flex;
    justify-content: space-between;
    gap: var(--space-2);
    padding: 2px var(--space-1);
    border-left: 3px solid transparent;
  }
  .row.warning {
    border-left-color: var(--color-warning);
  }
  .row.critical {
    border-left-color: var(--color-danger);
  }
  .row.warning dd {
    color: var(--color-warning);
  }
  .row.critical dd {
    color: var(--color-danger);
    font-weight: 600;
  }
  dt {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  dd {
    margin: 0;
    font-variant-numeric: tabular-nums;
    text-align: right;
    overflow-wrap: anywhere;
  }
  .small {
    font-size: 0.8em;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }
</style>
