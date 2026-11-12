<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatUptime } from '../../../shared/ui/format';
  import { type ArrOverview, parseWidgetConfig } from '../domain/arr';
  import type { ArrUseCases } from '../usecase/arrUseCases';

  /** Props of the Sonarr/Radarr/Prowlarr widget (#86). */
  interface Props {
    arr: ArrUseCases;
    config: Record<string, unknown>;
  }
  let { arr, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseWidgetConfig(config));

  const POLL_MS = 60_000;
  let overview = $state<ArrOverview | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  const day = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'short' });

  $effect(() =>
    poll(
      () =>
        arr
          .overview()
          .then((o) => ((overview = o), (error = ''), (unconfigured = false)))
          .catch((err) => {
            unconfigured = err instanceof ApiError && err.status === 404;
            error = errorMessage(err);
            throw err;
          }),
      POLL_MS,
    ),
  );
</script>

{#if unconfigured}
  <p class="muted">{t('arr.connect_under_settings')}</p>
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  {#if cfg.health && overview.health.length}
    <ul class="health">
      {#each overview.health as h, i (i)}
        <li class={h.severity}><strong>{h.instance}:</strong> {h.message}</li>
      {/each}
    </ul>
  {/if}
  {#if cfg.queue}
    <h4>{t('arr.queue')}</h4>
    {#each overview.queue as q, i (i)}
      <div class="queue" class:problem={q.error !== ''}>
        <span class="title" title={q.title}>{q.title}</span>
        <span class="muted small"
          >{q.error ||
            (q.timeLeftSeconds > 0
              ? t('arr.time_left', { value: formatUptime(q.timeLeftSeconds) })
              : q.status)}</span
        >
        <div
          class="bar"
          role="progressbar"
          aria-label={q.title}
          aria-valuenow={q.progress}
          aria-valuemin="0"
          aria-valuemax="100"
        >
          <div class="fill" style:width="{q.progress}%"></div>
        </div>
      </div>
    {:else}
      <p class="muted">{t('arr.queue_empty')}</p>
    {/each}
  {/if}
  {#if cfg.upcoming}
    <h4>{t('arr.upcoming')}</h4>
    <ul class="upcoming">
      {#each overview.upcoming as r, i (i)}
        <li>
          <span class="muted small date">{day(r.date)}</span>
          <span class="title" title="{r.title} · {r.detail}">{r.hasFile ? '✓ ' : ''}{r.title}</span>
          <span class="muted small">{r.detail}</span>
        </li>
      {:else}
        <li class="muted">{t('arr.nothing_upcoming')}</li>
      {/each}
    </ul>
    {#each Object.entries(overview.missing) as [name, n] (name)}
      <p class="muted small">{t('arr.missing', { name, count: n })}</p>
    {/each}
  {/if}
  {#each overview.errors as e (e)}<p class="error small">{t('arr.unreachable', { detail: e })}</p>{/each}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  h4 {
    margin: var(--space-2) 0 var(--space-1);
  }
  .health {
    list-style: none;
    padding: 0;
    margin: 0;
    font-size: 0.85em;
  }
  .health .error {
    color: var(--color-danger);
  }
  .health .warning {
    color: var(--color-warning);
  }
  .queue {
    display: grid;
    gap: 2px;
    margin-bottom: var(--space-2);
  }
  .queue.problem .small {
    color: var(--color-danger);
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
  }
  .bar {
    height: 4px;
    background: var(--color-border);
    border-radius: 2px;
  }
  .fill {
    height: 100%;
    background: var(--color-accent);
    border-radius: 2px;
  }
  .upcoming {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    gap: var(--space-1);
  }
  .upcoming li {
    display: grid;
    grid-template-columns: 6.5em minmax(0, 1fr);
    column-gap: var(--space-2);
  }
  .upcoming li > :last-child {
    grid-column: 2;
  }
  .small {
    font-size: 0.8em;
  }
</style>
