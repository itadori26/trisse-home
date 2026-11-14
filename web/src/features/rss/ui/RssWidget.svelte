<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type FeedItem, itemMeta, parseRssConfig } from '../domain/feed';
  import type { RssUseCases } from '../usecase/rssUseCases';

  /** Props of the RSS widget. */
  interface Props {
    rss: RssUseCases;
    config: Record<string, unknown>;
    /** Read later (#99); the button shows only when saving is possible. */
    readLater?: { available(): Promise<boolean>; save(url: string): Promise<unknown> };
  }
  let { rss, config, readLater }: Props = $props();
  let canReadLater = $state(false);
  let savedLater = $state<string[]>([]);
  $effect(() => {
    void readLater?.available().then((ok) => (canReadLater = ok));
  });
  async function saveForLater(item: FeedItem) {
    if (!readLater || !item.link) return;
    try {
      await readLater.save(item.link);
      savedLater = [...savedLater, item.id];
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const POLL_MS = 5 * 60 * 1000;
  /** Poll faster while the list is empty, e.g. right after feeds were added. */
  const EMPTY_POLL_MS = 15_000;
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
    let timer: ReturnType<typeof setTimeout>;
    const tick = async () => {
      await load();
      timer = setTimeout(tick, items.length === 0 ? EMPTY_POLL_MS : POLL_MS);
    };
    tick();
    return () => clearTimeout(timer);
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
          title={cfg.compact ? itemMeta(item, !!cfg.feedId, now) : undefined}
          onclick={() => open(item)}
          onauxclick={() => open(item)}>{item.title}</a
        >
      {:else}
        <button
          class="linklike"
          type="button"
          title={cfg.compact ? itemMeta(item, !!cfg.feedId, now) : undefined}
          onclick={() => open(item)}>{item.title}</button
        >
      {/if}
      {#if !cfg.compact}<span class="meta muted">{itemMeta(item, !!cfg.feedId, now)}</span>{/if}
      {#if canReadLater && item.link}
        <button
          class="btn btn-icon later"
          type="button"
          disabled={savedLater.includes(item.id)}
          title={savedLater.includes(item.id) ? t('readlater.saved_for_later') : t('readlater.read_later')}
          aria-label={t('readlater.read_later_item', { title: item.title })}
          onclick={() => saveForLater(item)}>{savedLater.includes(item.id) ? '✓' : '🔖'}</button
        >
      {/if}
    </li>
  {:else}
    <li class="muted">{cfg.unreadOnly ? t('rss.all_caught_up') : t('rss.no_items')}</li>
  {/each}
</ul>
{#if items.some((i) => !i.read)}
  <button class="btn btn-icon muted" type="button" onclick={markAll}>{t('rss.mark_all_as_read')}</button>
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
    font-size: 0.75em;
  }
  .later {
    padding: 0 var(--space-1);
    font-size: 0.8em;
  }
</style>
