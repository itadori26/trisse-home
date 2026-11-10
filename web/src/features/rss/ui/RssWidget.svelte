<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { age, type FeedItem, parseRssConfig } from '../domain/feed';
  import type { RssUseCases } from '../usecase/rssUseCases';

  /** Props of the RSS widget. */
  interface Props {
    rss: RssUseCases;
    config: Record<string, unknown>;
  }
  let { rss, config }: Props = $props();

  const POLL_MS = 5 * 60 * 1000;
  const cfg = $derived(parseRssConfig(config));
  let items = $state<FeedItem[]>([]);
  let error = $state('');
  let now = $state(new Date());

  async function load() {
    try {
      items = await rss.items(cfg);
      now = new Date();
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void cfg;
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  async function open(item: FeedItem) {
    const updated = await rss.open(item).catch(() => item);
    items = cfg.unreadOnly
      ? items.filter((i) => i.id !== item.id)
      : items.map((i) => (i.id === item.id ? updated : i));
  }

  async function markAll() {
    await rss.markAllRead(cfg.feedId).catch((err) => (error = errorMessage(err)));
    load();
  }
</script>

{#if error}<p class="error">{error}</p>{/if}
<ul class="list">
  {#each items as item (item.id)}
    <li class:read={item.read}>
      {#if item.link}
        <a
          href={item.link}
          target="_blank"
          rel="noopener noreferrer"
          onclick={() => open(item)}
          onauxclick={() => open(item)}>{item.title}</a
        >
      {:else}
        <button class="linklike" type="button" onclick={() => open(item)}>{item.title}</button>
      {/if}
      <span class="meta muted">{cfg.feedId ? '' : `${item.feedTitle} · `}{age(item.publishedAt, now)}</span>
    </li>
  {:else}
    <li class="muted">{cfg.unreadOnly ? 'All caught up.' : 'No items. Add feeds under Settings → Feeds.'}</li>
  {/each}
</ul>
{#if items.some((i) => !i.read)}
  <button class="btn btn-icon muted" type="button" onclick={markAll}>Mark all as read</button>
{/if}

<style>
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    padding: 4px 0;
    border-bottom: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
  }
  a,
  .linklike {
    color: inherit;
    text-decoration: none;
    font-weight: 500;
  }
  a:hover {
    text-decoration: underline;
  }
  .linklike {
    border: none;
    background: none;
    padding: 0;
    text-align: left;
    cursor: pointer;
  }
  .read a,
  .read .linklike {
    font-weight: 400;
    color: var(--color-text-muted);
  }
  .meta {
    font-size: 0.75rem;
  }
</style>
