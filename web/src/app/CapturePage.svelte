<script lang="ts">
  import { t } from '../shared/i18n';
  import { errorMessage } from '../shared/api-client/http';
  import {
    CAPTURE_TARGETS,
    type CaptureForm,
    type CaptureTarget,
    initialForm,
    parseShared,
  } from '../features/capture/domain/capture';
  import type { QuickResult } from '../features/quickactions/usecase/quickActionsUseCases';
  import { captureCommand } from './capture';
  import type { Container } from './container';

  /**
   * /capture (#66): saves what the bookmarklet, the share target or the
   * browser extension hands over. Opening the page changes nothing; only
   * the explicit Save does.
   */
  interface Props {
    container: Container;
  }
  let { container }: Props = $props();

  let form = $state<CaptureForm>(initialForm(parseShared(location.search)));
  let error = $state('');
  let busy = $state(false);
  let saved = $state<QuickResult | null>(null);
  let undone = $state(false);
  let title = $state<HTMLInputElement>();

  $effect(() => title?.focus());

  const label: Record<CaptureTarget, Parameters<typeof t>[0]> = {
    bookmark: 'capture.bookmark',
    todo: 'capture.todo',
    note: 'capture.note',
  };

  async function save(e: SubmitEvent) {
    e.preventDefault();
    const command = captureCommand(form, new Date());
    if (!command) {
      error = t('capture.missing');
      return;
    }
    busy = true;
    error = '';
    try {
      saved = await container.quickActions.execute(command);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function undo() {
    await saved?.undo?.();
    undone = true;
  }

  // Opened by the bookmarklet or the extension as a popup: it can close itself.
  const closable = window.opener !== null || history.length <= 1;
</script>

<main class="capture">
  <h1>{t('capture.title')}</h1>
  {#if saved}
    <p role="status">
      {undone ? t('capture.undone') : t(`capture.saved_${form.target}` as Parameters<typeof t>[0])}
    </p>
    <div class="actions">
      {#if saved.undo && !undone}<button class="btn" type="button" onclick={undo}>{t('capture.undo')}</button
        >{/if}
      {#if closable}<button class="btn btn-primary" type="button" onclick={() => window.close()}
          >{t('capture.close')}</button
        >{/if}
      <a class="btn" href="/">{t('capture.dashboard')}</a>
      <button class="btn" type="button" onclick={() => ((saved = null), (undone = false))}
        >{t('capture.another')}</button
      >
    </div>
  {:else}
    <form onsubmit={save}>
      <div class="targets" role="radiogroup" aria-label={t('capture.as')}>
        {#each CAPTURE_TARGETS as target (target)}
          <label class:active={form.target === target}
            ><input type="radio" bind:group={form.target} value={target} />{t(label[target])}</label
          >
        {/each}
      </div>
      <label class="field"
        ><span>{t('capture.name')}</span><input
          class="input"
          bind:this={title}
          bind:value={form.title}
          maxlength="200"
        /></label
      >
      {#if form.target !== 'todo'}
        <label class="field"
          ><span>{t('capture.url')}</span><input class="input" type="url" bind:value={form.url} /></label
        >
      {/if}
      {#if form.target === 'note'}
        <label class="field"
          ><span>{t('capture.text')}</span><textarea class="input" rows="5" bind:value={form.text}
          ></textarea></label
        >
      {/if}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <button class="btn btn-primary" type="submit" disabled={busy}>{t('capture.save')}</button>
    </form>
  {/if}
</main>

<style>
  .capture {
    max-width: 32rem;
    margin: 0 auto;
    padding: var(--space-4, 1.5rem) var(--space-3);
  }
  h1 {
    font-size: 1.3em;
    margin: 0 0 var(--space-3);
  }
  form {
    display: grid;
    gap: var(--space-2);
  }
  .targets {
    display: inline-flex;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    overflow: hidden;
    justify-self: start;
  }
  .targets label {
    padding: var(--space-1) var(--space-3);
    cursor: pointer;
  }
  .targets label.active {
    background: var(--color-accent);
    color: var(--color-accent-contrast);
  }
  .targets label:has(:focus-visible) {
    outline: 2px solid var(--color-accent);
    outline-offset: -2px;
  }
  .targets input {
    position: absolute;
    opacity: 0;
    width: 1px;
    height: 1px;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  button[type='submit'] {
    justify-self: start;
  }
</style>
