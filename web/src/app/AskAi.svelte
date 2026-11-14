<script lang="ts">
  import { t } from '../shared/i18n';
  import { errorMessage } from '../shared/api-client/http';

  /** "Ask the AI" in the search (#162): explicit, streamed, cancellable. Loaded lazily. */
  interface Props {
    question: string;
    ask: (onChunk: (text: string) => void, signal: AbortSignal) => Promise<void>;
  }
  let { question, ask }: Props = $props();

  let answer = $state('');
  let error = $state('');
  let controller = $state<AbortController | null>(null);
  let asked = $state('');

  async function start() {
    controller?.abort();
    const c = new AbortController();
    controller = c;
    answer = error = '';
    asked = question;
    try {
      await ask((chunk) => (answer += chunk), c.signal);
    } catch (err) {
      if (!c.signal.aborted) error = errorMessage(err);
    } finally {
      if (controller === c) controller = null;
    }
  }

  // A changed question drops the old answer.
  $effect(() => {
    if (question !== asked && asked) {
      controller?.abort();
      answer = error = asked = '';
    }
  });
  $effect(() => () => controller?.abort());
</script>

<section class="ai" aria-label={t('ai_ask.title')}>
  {#if !asked}
    <button class="btn" type="button" onclick={start}>✨ {t('ai_ask.button')}</button>
    <span class="muted small">{t('ai_ask.hint')}</span>
  {:else}
    <p class="label muted small">✨ {t('ai_ask.generated')}</p>
    <p class="answer" aria-live="polite" aria-busy={!!controller}>
      {answer}{#if controller && !answer}{t('ai_ask.thinking')}{/if}
    </p>
    {#if error}<p class="error small" role="alert">{error}</p>{/if}
    {#if controller}
      <button class="btn" type="button" onclick={() => controller?.abort()}>{t('ai_ask.stop')}</button>
    {:else}
      <button class="btn" type="button" onclick={start}>{t('ai_ask.again')}</button>
    {/if}
  {/if}
</section>

<style>
  .ai {
    padding: var(--space-2) var(--space-3);
    border-top: 1px solid var(--color-border);
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
  }
  .label,
  .answer {
    flex-basis: 100%;
    margin: 0;
  }
  .answer {
    white-space: pre-wrap;
  }
  .small {
    font-size: 0.8em;
  }
</style>
