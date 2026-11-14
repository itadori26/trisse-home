<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { type Endpoint, parseWidgetConfig } from '../domain/apiwidget';
  import type { ApiWidgetUseCases } from '../usecase/apiWidgetUseCases';

  /** Widget settings: which endpoint the widget shows. */
  interface Props {
    apiwidget: ApiWidgetUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { apiwidget, config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let endpointId = $state(parseWidgetConfig(config).endpointId);
  let endpoints = $state<Endpoint[] | null>(null);

  $effect(() => {
    apiwidget.endpoints().then((list) => {
      endpoints = list;
      if (!endpointId && list[0]) endpointId = list[0].id;
    });
  });
</script>

{#if endpoints && endpoints.length === 0}
  <p class="muted">{t('apiwidget.set_up')}</p>
{:else}
  <form
    onsubmit={(e) => {
      e.preventDefault();
      onsave({ endpointId });
    }}
  >
    <label class="field"
      ><span>{t('apiwidget.choose_endpoint')}</span><select class="input" bind:value={endpointId} required>
        {#each endpoints ?? [] as e (e.id)}<option value={e.id}>{e.name}</option>{/each}
      </select></label
    >
    <p><button class="btn btn-primary" type="submit" disabled={!endpointId}>{t('common.save')}</button></p>
  </form>
{/if}
