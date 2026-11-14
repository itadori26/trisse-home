<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    entryName,
    type HolidaysOverview,
    parseHolidaysConfig,
    periodDays,
    stateName,
  } from '../domain/holidays';
  import type { HolidaysUseCases } from '../usecase/holidaysUseCases';

  /** Props of the holidays widget (#55). */
  interface Props {
    holidays: HolidaysUseCases;
    config: Record<string, unknown>;
  }
  let { holidays, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseHolidaysConfig(config));

  // The list changes at midnight at the latest; an hourly look is plenty.
  const POLL_MS = 60 * 60_000;
  let overview = $state<HolidaysOverview | null>(null);
  let error = $state('');

  $effect(() => {
    const { state, count } = cfg;
    return poll(
      () =>
        holidays
          .upcoming(state, count)
          .then((o) => ((overview = o), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      POLL_MS,
    );
  });

  const date = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'short' });
  const when = (daysUntil: number) =>
    daysUntil === 0
      ? t('holidays.today')
      : daysUntil === 1
        ? t('holidays.tomorrow')
        : t('holidays.in_days', { n: daysUntil });
</script>

{#if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  <p class="state muted">{stateName(overview.state)}</p>
  <ul>
    {#each overview.entries as e (e.kind + e.key + e.name + e.start.getTime())}
      <li class:current={e.current} class:school={e.kind === 'school'}>
        <span class="what">
          <span class="name">{entryName(e)}</span>
          <span class="muted small">
            {#if e.kind === 'school'}
              {t('holidays.school')} · {date(e.start)} – {date(e.end)} · {t('holidays.days', {
                n: periodDays(e),
              })}
            {:else}
              {date(e.start)}
            {/if}
          </span>
        </span>
        <span class="when">
          {#if e.current && e.kind === 'school'}{t('holidays.until', { date: date(e.end) })}{:else}{when(
              e.daysUntil,
            )}{/if}
        </span>
      </li>
    {:else}
      <li class="muted">{t('holidays.none')}</li>
    {/each}
  </ul>
  {#if overview.schoolError}<p class="error small">{overview.schoolError}</p>{/if}
{/if}

<style>
  .state {
    margin: 0 0 var(--space-1);
    font-size: 0.8em;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 3px;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 2px var(--space-1);
    border-left: 3px solid transparent;
    border-radius: var(--radius-small);
  }
  li.school {
    border-left-color: color-mix(in srgb, var(--color-accent) 50%, transparent);
  }
  li.current {
    border-left-color: var(--color-accent);
    background: color-mix(in srgb, var(--color-accent) 12%, transparent);
  }
  .what {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .name,
  .small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .when {
    flex: none;
    font-size: 0.85em;
    color: var(--color-accent);
    font-variant-numeric: tabular-nums;
  }
  .current .when {
    font-weight: 600;
  }
  .small {
    font-size: 0.8em;
  }
</style>
