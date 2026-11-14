<script lang="ts">
  import { t } from '../i18n';
  import { errorMessage } from '../api-client/http';
  import type { SummaryOutput } from './summarizer';

  /** "Summarise" button with the result (#75); the caller decides what is summarised. */
  interface Props {
    run: () => Promise<SummaryOutput>;
    /** Accessible name of the button, e.g. "Summarise <title>". */
    label?: string;
    compact?: boolean;
  }
  let { run, label, compact = false }: Props = $props();

  let result = $state<SummaryOutput | null>(null);
  let busy = $state(false);
  let error = $state('');

  async function summarise() {
    if (result) {
      result = null; // the button toggles the summary
      return;
    }
    busy = true;
    error = '';
    try {
      result = await run();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<button
  class="btn"
  class:btn-icon={compact}
  type="button"
  disabled={busy}
  aria-expanded={!!result}
  aria-label={label ?? t('ai_summary.button')}
  title={t('ai_summary.button')}
  onclick={summarise}
  >✨{#if !compact}&nbsp;{busy ? t('ai_summary.working') : t('ai_summary.button')}{/if}</button
>
{#if busy && compact}<span class="muted small" role="status">{t('ai_summary.working')}</span>{/if}
{#if result}
  <div class="summary" role="note" aria-label={t('ai_summary.title')}>
    <p>{result.summary}</p>
    {#if result.truncated}<p class="muted small">{t('ai_summary.truncated')}</p>{/if}
  </div>
{/if}
{#if error}<p class="error small" role="alert">{error}</p>{/if}

<style>
  .summary {
    margin: var(--space-1) 0;
    padding: var(--space-2);
    border-left: 3px solid var(--color-accent);
    background: var(--color-surface-2);
    border-radius: var(--radius-small);
    font-size: 0.9em;
  }
  .summary p {
    margin: 0;
    white-space: pre-wrap;
  }
  .summary p + p {
    margin-top: var(--space-1);
  }
  .small {
    font-size: 0.8em;
  }
</style>
