<script lang="ts">
  import { type Feed, parseRssConfig } from '../domain/feed';
  import type { RssUseCases } from '../usecase/rssUseCases';

  /** Props of the RSS widget settings. */
  interface Props {
    rss: RssUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { rss, config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parseRssConfig(config));
  let feeds = $state<Feed[]>([]);
  $effect(() => {
    rss.feeds().then((f) => (feeds = f));
  });
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <label class="field">
    <span>Feed</span>
    <select class="input" bind:value={form.feedId}>
      <option value="">All feeds</option>
      {#each feeds as f (f.id)}<option value={f.id}>{f.title}</option>{/each}
    </select>
  </label>
  <label class="field"
    ><span>Items: {form.limit}</span><input
      type="range"
      min="5"
      max="100"
      step="5"
      bind:value={form.limit}
    /></label
  >
  <label><input type="checkbox" bind:checked={form.unreadOnly} /> Unread items only</label>
  <p><button class="btn btn-primary" type="submit">Save</button></p>
</form>
