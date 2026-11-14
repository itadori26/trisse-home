<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { SummaryUseCases } from '../usecase/summaryUseCases';

  /** Settings of the AI summaries (#75): a local Ollama and its model. */
  interface Props {
    summary: SummaryUseCases;
  }
  let { summary }: Props = $props();

  let url = $state('');
  let model = $state('');
  let maxInputChars = $state(12000);
  let configured = $state(false);
  let models = $state<string[]>([]);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    summary
      .settings()
      .then((s) => {
        configured = s.configured;
        url = s.url;
        model = s.model;
        maxInputChars = s.maxInputChars;
        if (s.model) models = [s.model];
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function run(action: () => Promise<void>) {
    error = message = '';
    busy = true;
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  const loadModels = () =>
    run(async () => {
      models = await summary.models(url);
      if (!models.includes(model)) model = models[0] ?? '';
      message = models.length ? t('summary.models_found', { n: models.length }) : t('summary.no_models');
    });
</script>

<p class="muted">{t('summary.hint')}</p>
<form
  onsubmit={(e) => {
    e.preventDefault();
    void run(async () => {
      const s = await summary.saveSettings(url, model, maxInputChars);
      configured = s.configured;
      message = t('summary.saved');
    });
  }}
>
  <label class="field" for="summary-url"><span>{t('summary.url')}</span></label>
  <div class="row url">
    <input
      id="summary-url"
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://192.168.1.10:11434"
      required
    />
    <button class="btn" type="button" disabled={busy || !url} onclick={loadModels}
      >{t('summary.load_models')}</button
    >
  </div>
  <label class="field">
    <span>{t('summary.model')}</span>
    <select class="input" bind:value={model} required>
      {#each models as m (m)}<option value={m}>{m}</option>{/each}
    </select>
  </label>
  <label class="field">
    <span>{t('summary.max_input', { n: maxInputChars.toLocaleString() })}</span>
    <input type="range" min="1000" max="50000" step="1000" bind:value={maxInputChars} />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted" role="status">{message}</p>{/if}
  <div class="row">
    <button class="btn btn-primary" type="submit" disabled={busy || !model}>{t('summary.save')}</button>
    {#if configured}
      <button
        class="btn"
        type="button"
        disabled={busy}
        onclick={() =>
          run(async () => {
            await summary.disable();
            configured = false;
            message = t('summary.disabled');
          })}>{t('summary.disable')}</button
      >
    {/if}
  </div>
</form>

<style>
  .row {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .url {
    margin-bottom: var(--space-3);
  }
  .row .input {
    flex: 1;
    min-width: 12rem;
  }
</style>
