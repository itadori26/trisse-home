<script lang="ts">
  import { type Freshness, minutesAgo } from '../freshness/status';
  import { intlLocale, t } from '../i18n';

  /** Props of the shared data freshness badge (#94). */
  interface Props {
    freshness: Freshness;
    fetchedAt: number | null;
    lastError: string | null;
    now: number;
    onrefresh: () => void;
  }
  let { freshness, fetchedAt, lastError, now, onrefresh }: Props = $props();

  const age = $derived.by(() => {
    if (fetchedAt === null) return t('freshness.not_loaded');
    const m = minutesAgo(fetchedAt, now);
    return m === 0 ? t('freshness.just_now') : t('freshness.minutes_ago', { minutes: m });
  });
  const title = $derived(
    [
      fetchedAt === null
        ? ''
        : t('freshness.updated_at', { time: new Date(fetchedAt).toLocaleString(intlLocale()) }),
      lastError ? t('freshness.last_error', { error: lastError }) : '',
    ]
      .filter(Boolean)
      .join('\n'),
  );
</script>

<div class="badge {freshness}" title={title || undefined}>
  {#if freshness === 'error'}<span aria-hidden="true">⚠</span>{/if}
  <span class="age">
    {freshness === 'stale'
      ? t('freshness.stale', { age })
      : freshness === 'error'
        ? t('freshness.failed', { age })
        : age}
  </span>
  <button
    class="refresh"
    type="button"
    onclick={onrefresh}
    aria-label={t('freshness.refresh')}
    title={t('freshness.refresh')}>↻</button
  >
</div>

<style>
  .badge {
    position: absolute;
    right: 4px;
    bottom: 2px;
    display: flex;
    gap: 4px;
    align-items: center;
    font-size: 0.7rem;
    color: var(--color-text-muted);
    /* visibility (not only opacity) keeps the hidden badge away from screen
       readers; it switches at the end of the fade-out. */
    opacity: 0;
    visibility: hidden;
    transition:
      opacity 0.15s,
      visibility 0.15s;
    pointer-events: none;
  }
  /* Shown on hover/focus; stale data and errors are always visible. */
  :global(.lazy-root:hover) .badge,
  :global(.lazy-root:focus-within) .badge,
  .badge.stale,
  .badge.error {
    opacity: 1;
    visibility: visible;
    pointer-events: auto;
  }
  .badge.stale {
    color: var(--color-warning);
  }
  .badge.error {
    color: var(--color-danger);
  }
  .refresh {
    border: 0;
    background: none;
    color: inherit;
    cursor: pointer;
    padding: 0 2px;
    font: inherit;
  }
  @media (prefers-reduced-motion: reduce) {
    .badge {
      transition: none;
    }
  }
</style>
