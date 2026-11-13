<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type DropItem, EXPIRIES, type ExpiryHours, expiresIn, formatSize } from '../domain/drop';
  import type { DropUseCases } from '../usecase/dropUseCases';

  /** The drop zone (#101): paste text or links, drop files, pick them up elsewhere. */
  interface Props {
    drop: DropUseCases;
  }
  let { drop }: Props = $props();
  const poll = usePolling();

  let items = $state<DropItem[]>([]);
  let text = $state('');
  let expiry = $state<ExpiryHours>(24);
  let error = $state('');
  let busy = $state(false);
  let over = $state(false);
  let copied = $state('');
  let now = $state(new Date());
  let picker = $state<HTMLInputElement>();

  // Another device may have dropped something: a look every half minute.
  const load = () =>
    drop
      .list()
      .then((i) => ((items = i), (now = new Date()), (error = '')))
      .catch((err) => {
        error = errorMessage(err);
        throw err;
      });
  $effect(() => poll(load, 30_000));

  async function run(action: () => Promise<unknown>) {
    error = '';
    busy = true;
    try {
      await action();
      await load();
      return true;
    } catch (err) {
      error = errorMessage(err);
      return false;
    } finally {
      busy = false;
    }
  }

  const files = (list: FileList | null | undefined) => run(() => drop.addFiles([...(list ?? [])], expiry));

  async function copy(item: DropItem) {
    await navigator.clipboard?.writeText(item.text);
    copied = item.id;
    setTimeout(() => copied === item.id && (copied = ''), 1500);
  }
</script>

<div
  class="zone"
  class:over
  role="group"
  aria-label={t('drop.zone')}
  ondragover={(e) => (e.preventDefault(), (over = true))}
  ondragleave={() => (over = false)}
  ondrop={(e) => (e.preventDefault(), (over = false), void files(e.dataTransfer?.files))}
>
  <form
    onsubmit={async (e) => {
      e.preventDefault();
      if (await run(() => drop.addText(text, expiry))) text = '';
    }}
  >
    <textarea
      class="input"
      rows="2"
      bind:value={text}
      placeholder={t('drop.placeholder')}
      aria-label={t('drop.text')}></textarea>
    <div class="row">
      <select class="input" bind:value={expiry} aria-label={t('drop.expiry')}>
        {#each EXPIRIES as h (h)}<option value={h}>{t(`drop.expiry_${h}`)}</option>{/each}
      </select>
      <label class="btn"
        >{t('drop.file')}<input
          class="file"
          bind:this={picker}
          type="file"
          multiple
          disabled={busy}
          onchange={() => files(picker?.files).then(() => picker && (picker.value = ''))}
        /></label
      >
      <button class="btn btn-primary" type="submit" disabled={busy || !text.trim()}>{t('drop.drop')}</button>
    </div>
  </form>
</div>
{#if error}<p class="error small" role="alert">{error}</p>{/if}
<ul class="list">
  {#each items as i (i.id)}
    <li>
      <span class="what">
        {#if i.kind === 'link'}
          <a href={i.text} target="_blank" rel="noopener noreferrer" class="text">{i.text}</a>
        {:else if i.kind === 'file'}
          <span class="text">📄 {i.fileName}</span>
        {:else}
          <span class="text">{i.text}</span>
        {/if}
        <span class="muted small">
          {#if i.kind === 'file'}{formatSize(i.size, intlLocale())} ·
          {/if}{t('drop.expires', { when: expiresIn(i, now) })}
        </span>
      </span>
      {#if i.kind === 'file'}
        <a
          class="btn btn-icon"
          href={drop.fileUrl(i.id)}
          download={i.fileName}
          aria-label={t('drop.download', { name: i.fileName })}>⬇</a
        >
      {:else}
        <button class="btn btn-icon" type="button" aria-label={t('drop.copy')} onclick={() => copy(i)}
          >{copied === i.id ? '✓' : '⧉'}</button
        >
      {/if}
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('drop.delete')}
        onclick={() => run(() => drop.remove(i.id))}>✕</button
      >
    </li>
  {:else}
    <li class="muted">{t('drop.empty')}</li>
  {/each}
</ul>

<style>
  .zone {
    border: 2px dashed transparent;
    border-radius: var(--radius-small);
  }
  .zone.over {
    border-color: var(--color-accent);
    background: color-mix(in srgb, var(--color-accent) 8%, transparent);
  }
  form {
    display: grid;
    gap: var(--space-1);
  }
  textarea {
    resize: vertical;
  }
  .row {
    display: flex;
    gap: var(--space-1);
    flex-wrap: wrap;
  }
  .row select {
    flex: 1;
    min-width: 6rem;
  }
  .file {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
  }
  .row label:has(:focus-visible) {
    outline: 2px solid var(--color-accent);
  }
  .list {
    list-style: none;
    margin: var(--space-2) 0 0;
    padding: 0;
    display: grid;
    gap: 4px;
  }
  li {
    display: flex;
    align-items: flex-start;
    gap: var(--space-1);
  }
  .what {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8em;
  }
</style>
