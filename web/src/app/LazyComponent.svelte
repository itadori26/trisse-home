<script lang="ts">
  import { t } from '../shared/i18n';
  import type { Renderer } from './widgets';

  /** Props: a lazily loaded component and its props (#116). */
  interface Props {
    renderer: Renderer;
    props: Record<string, unknown>;
  }
  let { renderer, props }: Props = $props();

  // Re-created on retry; widgets on the same tab load in parallel because
  // every instance starts its (cached) loader at once.
  let attempt = $state(0);
  const loading = $derived.by(() => {
    void attempt; // dependency: a retry loads again
    return renderer.load();
  });
</script>

{#await loading}
  <div class="skeleton" aria-busy="true" aria-label={t('app.loading_widget')}></div>
{:then Component}
  <Component {...props} />
{:catch}
  <p class="error">
    {t('app.widget_failed_to_load')}
    <button class="btn" type="button" onclick={() => attempt++}>{t('app.retry')}</button>
  </p>
{/await}

<style>
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
