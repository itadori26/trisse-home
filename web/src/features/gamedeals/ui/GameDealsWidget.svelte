<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { GameDealsOverview } from '../domain/gamedeals';
  import type { GamedealsUseCases } from '../usecase/gamedealsUseCases';

  /** Props of the game deals widget (#62). */
  interface Props {
    gamedeals: GamedealsUseCases;
  }
  let { gamedeals }: Props = $props();
  const poll = usePolling();

  let overview = $state<GameDealsOverview | null>(null);
  let unconfigured = $state(false);
  let error = $state('');

  const date = (d: Date) => d.toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short' });

  $effect(() =>
    poll(
      () =>
        Promise.all([gamedeals.settings(), gamedeals.overview()])
          .then(([s, o]) => ((unconfigured = s.steamId === '' && !s.epic), (overview = o), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      60 * 60_000,
    ),
  );
</script>

{#if unconfigured}
  <p class="muted">{t('gamedeals.connect_under_settings')}</p>
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  {#if overview.freeNow.length || overview.freeNext.length}
    <h4>{t('gamedeals.free')}</h4>
    <ul class="list">
      {#each overview.freeNow as g (g.title)}
        <li>
          <a href={g.url} target="_blank" rel="noopener noreferrer">{g.title}</a>
          <span class="muted small">{t('gamedeals.until', { date: date(g.end) })}</span>
        </li>
      {/each}
      {#each overview.freeNext as g (g.title)}
        <li class="next">
          <a href={g.url} target="_blank" rel="noopener noreferrer">{g.title}</a>
          <span class="muted small">{t('gamedeals.from', { date: date(g.start) })}</span>
        </li>
      {/each}
    </ul>
  {/if}
  {#if overview.deals.length}
    <h4>{t('gamedeals.wishlist')}</h4>
    <ul class="list">
      {#each overview.deals as d (d.url)}
        <li>
          <span class="badge">−{d.discountPct} %</span>
          <a href={d.url} target="_blank" rel="noopener noreferrer">{d.name}</a>
          <span class="muted small"
            >{d.finalPrice}{d.endsAt ? ` · ${t('gamedeals.until', { date: date(d.endsAt) })}` : ''}</span
          >
        </li>
      {/each}
    </ul>
  {/if}
  {#if !overview.deals.length && !overview.freeNow.length && !overview.freeNext.length}
    <p class="muted">{t('gamedeals.nothing')}</p>
  {/if}
  {#each overview.errors as e (e)}<p class="error small">
      {t('gamedeals.unreachable', { source: e })}
    </p>{/each}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  h4 {
    margin: 0 0 var(--space-1);
  }
  .list {
    list-style: none;
    padding: 0;
    margin: 0 0 var(--space-2);
    display: grid;
    gap: var(--space-1);
  }
  .list li {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: var(--space-1) var(--space-2);
  }
  .next a {
    opacity: 0.75;
  }
  .badge {
    background: var(--color-success);
    color: var(--color-accent-contrast);
    border-radius: 4px;
    padding: 0 4px;
    font-size: 0.8em;
    font-weight: 600;
  }
  .small {
    font-size: 0.8em;
  }
</style>
