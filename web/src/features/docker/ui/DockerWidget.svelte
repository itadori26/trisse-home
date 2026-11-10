<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { type ContainerList, filterContainers, indicator, parseDockerConfig } from '../domain/container';
  import type { DockerUseCases } from '../usecase/dockerUseCases';

  /** Props of the docker widget. */
  interface Props {
    docker: DockerUseCases;
    config: Record<string, unknown>;
  }
  let { docker, config }: Props = $props();

  const POLL_MS = 20_000;
  const cfg = $derived(parseDockerConfig(config));
  let data = $state<ContainerList | null>(null);
  let error = $state('');
  const shown = $derived(data ? filterContainers(data.containers, cfg) : []);

  $effect(() => {
    const load = () =>
      docker
        .containers()
        .then((d) => ((data = d), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });
</script>

{#if error}<p class="error">{error}</p>{/if}
{#if data}
  <p class="summary">
    <span><strong>{data.running}</strong> running</span>
    <span class:alert={data.stopped > 0}><strong>{data.stopped}</strong> stopped</span>
    {#if data.unhealthy}<span class="alert"><strong>{data.unhealthy}</strong> unhealthy</span>{/if}
  </p>
  {#if data.error}<p class="error small">{data.error}</p>{/if}
  <ul class="list">
    {#each shown as c (c.id)}
      <li title="{c.image}\n{c.status}">
        <span class="dot {indicator(c)}" aria-label={c.state}></span>
        <span class="name">{c.name}</span>
        <span class="muted small status">{c.status}</span>
      </li>
    {:else}
      <li class="muted">{cfg.problemsOnly ? 'All containers are healthy. 👍' : 'No containers.'}</li>
    {/each}
  </ul>
{/if}

<style>
  .summary {
    display: flex;
    gap: var(--space-3);
    margin: 0 0 var(--space-2);
    font-size: 0.85rem;
  }
  .alert {
    color: var(--color-danger);
  }
  .small {
    font-size: 0.75rem;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 2px 0;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .status {
    white-space: nowrap;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex: none;
  }
  .ok {
    background: var(--color-success);
  }
  .warning {
    background: var(--color-warning);
  }
  .error.dot {
    background: var(--color-danger);
  }
  .off {
    background: var(--color-text-muted);
  }
</style>
