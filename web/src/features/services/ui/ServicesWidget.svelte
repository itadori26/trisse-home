<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Service, statusText } from '../domain/service';
  import type { ServicesUseCases } from '../usecase/servicesUseCases';
  import ServiceIcon from './ServiceIcon.svelte';

  /** Props of the service tiles widget. */
  interface Props {
    services: ServicesUseCases;
    config: Record<string, unknown>;
  }
  let { services, config }: Props = $props();

  const POLL_MS = 30_000;
  let list = $state<Service[]>([]);
  let error = $state('');
  const compact = $derived(config['compact'] === true);

  $effect(() => {
    const load = () =>
      services
        .list()
        .then((l) => ((list = l), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });
</script>

{#if error}<p class="error">{error}</p>{/if}
{#if list.length === 0 && !error}
  <p class="muted">No services yet. Add them under Settings → Services.</p>
{/if}
<ul class="tiles" class:compact>
  {#each list as s (s.id)}
    <li>
      <a href={s.url} target="_blank" rel="noopener noreferrer" title={statusText(s)}>
        <ServiceIcon icon={s.icon} name={s.name} />
        <span class="text">
          <span class="name">{s.name}</span>
          {#if s.description && !compact}<span class="muted desc">{s.description}</span>{/if}
        </span>
        <span class="dot {s.state}" aria-label={s.state}></span>
      </a>
    </li>
  {/each}
</ul>

<style>
  .tiles {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: var(--space-2);
  }
  .tiles.compact {
    grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  }
  a {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2);
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
    color: inherit;
    text-decoration: none;
  }
  a:hover {
    outline: 1px solid var(--color-accent);
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
  }
  .name,
  .desc {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .desc {
    font-size: 0.8rem;
  }
  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex: none;
    background: var(--color-text-muted);
  }
  .dot.online {
    background: var(--color-success);
  }
  .dot.offline {
    background: var(--color-danger);
  }
</style>
