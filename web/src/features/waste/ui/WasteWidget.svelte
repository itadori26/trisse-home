<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { dueLabel, isDueSoon, type WasteOverview } from '../domain/waste';
  import type { WasteUseCases } from '../usecase/wasteUseCases';

  /** Props of the waste widget. */
  interface Props {
    waste: WasteUseCases;
  }
  let { waste }: Props = $props();

  const POLL_MS = 30 * 60_000;
  let overview = $state<WasteOverview | null>(null);
  let error = $state('');
  let unconfigured = $state(false);
  let now = $state(new Date());

  $effect(() => {
    const load = () =>
      waste
        .upcoming()
        .then((o) => ((overview = o), (now = new Date()), (error = ''), (unconfigured = false)))
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
        });
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  const date = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'numeric' });
</script>

{#if unconfigured}
  <p class="muted">{t('waste.add_the_ics_link_of_your')}</p>
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  <ul class="types">
    {#each overview.types as t (t.label)}
      <li class:soon={isDueSoon(t, now)} style="--type-color: {t.color}">
        <span class="icon" aria-hidden="true">{t.icon}</span>
        <span class="label">{t.label}</span>
        <span class="when">
          <strong>{dueLabel(t.daysUntil)}</strong>
          {#if t.dates[0] && t.daysUntil > 1}<span class="muted small">{date(t.dates[0])}</span>{/if}
        </span>
      </li>
    {:else}
      <li class="muted">{t('waste.no_upcoming_collections')}</li>
    {/each}
  </ul>
  {#if overview.error}<p class="error small">{overview.error}</p>{/if}
{:else}
  <p class="muted">{t('waste.loading')}</p>
{/if}

<style>
  .types {
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
    padding: 2px 6px;
    border-left: 4px solid var(--type-color);
    border-radius: var(--radius-small);
  }
  li.soon {
    background: color-mix(in srgb, var(--type-color) 22%, transparent);
    font-weight: 600;
  }
  .label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .when {
    display: flex;
    gap: var(--space-1);
    align-items: baseline;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8em;
  }
</style>
