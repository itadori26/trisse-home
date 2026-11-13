<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { displayTitle, type MediaOverview, parseWidgetConfig, posterUrl } from '../domain/media';
  import type { MediaUseCases } from '../usecase/mediaUseCases';

  /** Props of the media server widget (#61). */
  interface Props {
    media: MediaUseCases;
    config: Record<string, unknown>;
  }
  let { media, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseWidgetConfig(config));

  const POLL_MS = 30_000;
  let overview = $state<MediaOverview | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  $effect(() =>
    poll(
      () =>
        media
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
  <p class="muted">{t('media.connect_under_settings')}</p>
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  {#if cfg.showSessions}
    <h4>{t('media.now_playing')}</h4>
    {#each overview.sessions as s, i (i)}
      <div class="session">
        <span class="title">{s.paused ? '⏸' : '▶'} {displayTitle(s)}</span>
        <span class="muted small">{s.user} · {s.device}</span>
        <div
          class="bar"
          role="progressbar"
          aria-label={displayTitle(s)}
          aria-valuenow={s.progress}
          aria-valuemin="0"
          aria-valuemax="100"
        >
          <div class="fill" style:width="{s.progress}%"></div>
        </div>
      </div>
    {:else}
      <p class="muted">{t('media.nothing_playing')}</p>
    {/each}
  {/if}
  {#if cfg.showRecent && overview.recent.length}
    <h4>{t('media.recently_added')}</h4>
    <ul class="posters">
      {#each overview.recent as item (item.id)}
        <li title={displayTitle(item)}>
          <img src={posterUrl(item.id)} alt="" loading="lazy" />
          <span class="small">{item.series || item.title}</span>
        </li>
      {/each}
    </ul>
  {/if}
  {#if error}<p class="error small">{error}</p>{/if}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  h4 {
    margin: 0 0 var(--space-1);
  }
  .session {
    display: grid;
    gap: 2px;
    margin-bottom: var(--space-2);
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
  .posters {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(4.5rem, 1fr));
    gap: var(--space-2);
  }
  .posters li {
    display: grid;
    gap: 2px;
    min-width: 0;
  }
  .posters img {
    width: 100%;
    aspect-ratio: 2 / 3;
    object-fit: cover;
    border-radius: 4px;
    background: var(--color-border);
  }
  .posters span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8em;
  }
</style>
