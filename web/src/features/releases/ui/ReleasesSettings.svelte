<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Artist, artistLabel } from '../domain/releases';
  import type { ReleasesUseCases } from '../usecase/releasesUseCases';

  /** Props of the followed-artists settings. */
  interface Props {
    releases: ReleasesUseCases;
  }
  let { releases }: Props = $props();

  let followed = $state<Artist[]>([]);
  let query = $state('');
  let results = $state<Artist[]>([]);
  let error = $state('');
  let busy = $state(false);

  async function run(action: () => Promise<void>) {
    error = '';
    busy = true;
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  $effect(() => {
    run(async () => {
      followed = await releases.followed();
    });
  });

  function search(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      results = await releases.search(query);
    });
  }

  function follow(a: Artist) {
    run(async () => {
      await releases.follow(a);
      followed = await releases.followed();
      results = results.filter((r) => r.id !== a.id);
    });
  }

  function unfollow(a: Artist) {
    run(async () => {
      await releases.unfollow(a.id);
      followed = followed.filter((f) => f.id !== a.id);
    });
  }
</script>

<p class="muted">{t('releases.hint')}</p>
<form class="search" onsubmit={search}>
  <input
    class="input"
    bind:value={query}
    placeholder={t('releases.search')}
    aria-label={t('releases.search')}
  />
  <button class="btn" type="submit" disabled={busy}>{t('weather.search')}</button>
</form>
{#if results.length}
  <ul class="list">
    {#each results as a (a.id)}
      <li>
        <span class="name">{artistLabel(a)}</span>
        <button
          class="btn"
          type="button"
          disabled={busy || followed.some((f) => f.id === a.id)}
          onclick={() => follow(a)}>{t('releases.follow')}</button
        >
      </li>
    {/each}
  </ul>
{/if}
<h4>{t('releases.followed', { count: followed.length })}</h4>
<ul class="list">
  {#each followed as a (a.id)}
    <li>
      <span class="name">{a.name}</span>
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('releases.unfollow', { name: a.name })}
        onclick={() => unfollow(a)}>✕</button
      >
    </li>
  {:else}
    <li class="muted">{t('releases.no_artists')}</li>
  {/each}
</ul>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<p class="muted small">{t('releases.attribution')}</p>

<style>
  .search {
    display: flex;
    gap: var(--space-2);
  }
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .list li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .name {
    flex: 1;
  }
  .small {
    font-size: 0.8em;
  }
</style>
