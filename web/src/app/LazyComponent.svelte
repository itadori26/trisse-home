<script lang="ts">
  import { combine, type Freshness } from '../shared/freshness/status';
  import { t } from '../shared/i18n';
  import WidgetStatusBadge from '../shared/ui/WidgetStatusBadge.svelte';
  import {
    observeViewport,
    provideRefresh,
    type RefreshHandle,
    type RefreshScheduler,
  } from '../shared/refresh';
  import type { Renderer } from './widgets';

  /** Props: a lazily loaded component and its props (#116). */
  interface Props {
    renderer: Renderer;
    props: Record<string, unknown>;
    /** Shared scheduler; the wrapper reports the widget's viewport state (#117). */
    scheduler: RefreshScheduler;
    /** Shows the data freshness badge for polling widgets (#94). */
    showStatus?: boolean;
  }
  let { renderer, props, scheduler, showStatus = true }: Props = $props();

  // Polling jobs of the widget learn whether its card is on screen.
  const handles: RefreshHandle[] = []; // not reactive on purpose
  let onScreen = true;
  let root = $state<HTMLElement>();
  provideRefresh({
    get scheduler() {
      return scheduler;
    },
    track(h) {
      handles.push(h);
      h.setInViewport(onScreen);
      const off = h.onSettled(updateStatus);
      updateStatus();
      return () => {
        off();
        handles.splice(handles.indexOf(h) >>> 0, 1);
        updateStatus();
      };
    },
  });

  // Data freshness of the widget's polling jobs (#94).
  let now = $state(Date.now());
  let status = $state<{ freshness: Freshness; fetchedAt: number | null; lastError: string | null } | null>(
    null,
  );
  function updateStatus() {
    now = Date.now();
    status = handles.length
      ? combine(
          handles.map((h) => h.status()),
          now,
        )
      : null;
  }
  $effect(() => {
    if (!showStatus) return;
    // Keeps "x min ago" and the stale marker current; cheap, no requests.
    const id = setInterval(updateStatus, 30_000);
    return () => clearInterval(id);
  });
  $effect(() => {
    if (!root) return;
    return observeViewport(root, (v) => {
      onScreen = v;
      for (const h of handles) h.setInViewport(v);
    });
  });

  // Re-created on retry; widgets on the same tab load in parallel because
  // every instance starts its (cached) loader at once.
  let attempt = $state(0);
  const loading = $derived.by(() => {
    void attempt; // dependency: a retry loads again
    return renderer.load();
  });
</script>

<div class="lazy-root" bind:this={root}>
  {#await loading}
    <div class="skeleton" aria-busy="true" aria-label={t('app.loading_widget')}></div>
  {:then Component}
    <Component {...props} />
    {#if showStatus && status}
      <WidgetStatusBadge
        freshness={status.freshness}
        fetchedAt={status.fetchedAt}
        lastError={status.lastError}
        {now}
        onrefresh={() => handles.forEach((h) => h.refreshNow())}
      />
    {/if}
  {:catch}
    <p class="error">
      {t('app.widget_failed_to_load')}
      <button class="btn" type="button" onclick={() => attempt++}>{t('app.retry')}</button>
    </p>
  {/await}
</div>

<style>
  .lazy-root {
    position: relative;
    height: 100%;
  }
  .skeleton {
    height: 100%;
    min-height: 3rem;
    border-radius: var(--radius-small);
    background: linear-gradient(
      90deg,
      var(--color-surface-2) 25%,
      transparent 50%,
      var(--color-surface-2) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.4s infinite linear;
    opacity: 0.6;
  }
  @keyframes shimmer {
    to {
      background-position: -200% 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .skeleton {
      animation: none;
    }
  }
</style>
