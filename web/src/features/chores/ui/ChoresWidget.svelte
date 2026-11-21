<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { type Chore, daysUntil, splitDue } from '../domain/chores';
  import type { ChoresUseCases } from '../usecase/choresUseCases';
  import ChoreDialog from './ChoreDialog.svelte';

  /** Props of the chore plan widget (#683). */
  interface Props {
    chores: ChoresUseCases;
  }
  let { chores }: Props = $props();
  const poll = usePolling();

  let items = $state<Chore[]>([]);
  let error = $state('');
  let busy = $state('');
  let editing = $state<Chore | null | undefined>(undefined); // null = new

  const groups = $derived(splitDue(items));
  const todayIso = () => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  };

  async function load() {
    try {
      items = await chores.list();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => poll(load, 15 * 60_000));

  function when(c: Chore): string {
    const n = daysUntil(c.due, new Date());
    if (n < 0) return t('chores.overdue_days', { n: -n });
    if (n === 0) return t('chores.today');
    if (n === 1) return t('chores.tomorrow');
    return t('chores.in_days', { n });
  }

  async function done(c: Chore) {
    busy = c.id;
    try {
      await chores.done(c.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }

  async function remove(c: Chore) {
    if (!confirm(t('common.confirm_delete', { name: c.title }))) return;
    try {
      await chores.remove(c.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#snippet row(c: Chore)}
  <li class:overdue={c.overdue}>
    {#if c.dueNow}
      <button
        class="btn btn-icon tick"
        type="button"
        disabled={busy === c.id}
        aria-label={t('chores.done_named', { title: c.title })}
        onclick={() => done(c)}>✓</button
      >
    {/if}
    <span class="main">
      <span class="title">{c.title}</span>
      <span class="meta muted">
        {when(c)}{#if c.assignee}&nbsp;· {t('chores.turn', { name: c.assignee })}{/if}
      </span>
    </span>
    {#if c.mine}
      <span class="actions">
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('chores.edit_named', { title: c.title })}
          onclick={() => (editing = c)}>✎</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('chores.remove', { title: c.title })}
          onclick={() => remove(c)}>✕</button
        >
      </span>
    {/if}
  </li>
{/snippet}

{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if items.length === 0}
  <p class="muted">{t('chores.empty')}</p>
{:else}
  {#if groups.due.length > 0}
    <h3 class="heading">{t('chores.due_heading', { n: groups.due.length })}</h3>
    <ul>
      {#each groups.due as c (c.id)}{@render row(c)}{/each}
    </ul>
  {:else}
    <p class="muted">{t('chores.nothing_today')}</p>
  {/if}
  {#if groups.later.length > 0}
    <h3 class="heading">{t('chores.later_heading')}</h3>
    <ul>
      {#each groups.later as c (c.id)}{@render row(c)}{/each}
    </ul>
  {/if}
{/if}
<button class="btn add" type="button" onclick={() => (editing = null)}>+ {t('chores.new')}</button>

{#if editing !== undefined}
  <ChoreDialog
    {chores}
    chore={editing}
    today={todayIso()}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = undefined)}
  />
{/if}

<style>
  .heading {
    font-size: 0.8rem;
    font-weight: 600;
    margin: var(--space-1) 0;
    color: var(--color-text-muted);
  }
  ul {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 0;
  }
  .tick {
    min-width: 44px;
    min-height: 44px;
  }
  .main {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  .title,
  .meta {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meta {
    font-size: 0.8rem;
  }
  li.overdue .meta {
    color: var(--color-accent-text);
  }
  .actions {
    display: flex;
  }
</style>
