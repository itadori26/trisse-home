<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Feed } from '../domain/feed';
  import type { RssUseCases } from '../usecase/rssUseCases';

  /** Props of the feed management section. */
  interface Props {
    rss: RssUseCases;
  }
  let { rss }: Props = $props();

  let feeds = $state<Feed[]>([]);
  let url = $state('');
  let error = $state('');
  let busy = $state(false);

  async function run(action: () => Promise<Feed[] | unknown>) {
    error = '';
    busy = true;
    try {
      const result = await action();
      feeds = Array.isArray(result) ? (result as Feed[]) : await rss.feeds();
      return true;
    } catch (err) {
      error = errorMessage(err);
      return false;
    } finally {
      busy = false;
    }
  }

  run(() => rss.feeds());
</script>

<ul class="list">
  {#each feeds as f (f.id)}
    <li>
      <span class="name">
        <strong>{f.title}</strong>
        {#if f.unread}<span class="badge">{f.unread}</span>{/if}
        <br /><span class="muted small">{f.url}</span>
        {#if f.lastError}<br /><span class="error small">{f.lastError}</span>{/if}
      </span>
      <button
        class="btn btn-icon"
        type="button"
        aria-label="Unsubscribe {f.title}"
        onclick={() => confirm(`Unsubscribe from "${f.title}"?`) && run(() => rss.unsubscribe(f.id))}
        >✕</button
      >
    </li>
  {/each}
</ul>
<form
  class="add"
  onsubmit={async (e) => {
    e.preventDefault();
    if (await run(() => rss.subscribe(url))) url = '';
  }}
>
  <input
    class="input"
    type="url"
    bind:value={url}
    placeholder="https://example.com/feed.xml"
    required
    aria-label="Feed URL"
  />
  <button class="btn" type="submit" disabled={busy}>Subscribe</button>
</form>
<p>
  <button class="btn btn-icon" type="button" disabled={busy} onclick={() => run(() => rss.refresh())}
    >↻ Refresh all feeds now</button
  >
</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  .list li {
    display: flex;
    align-items: flex-start;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .small {
    font-size: 0.8rem;
  }
  .badge {
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    border-radius: 999px;
    padding: 0 6px;
    font-size: 0.75rem;
  }
  .add {
    display: flex;
    gap: var(--space-2);
  }
</style>
