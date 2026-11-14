<script lang="ts">
  import { t } from '../../shared/i18n';
  import Modal from '../../shared/ui/Modal.svelte';
  import type { Conflict } from '../../shared/offline/domain/outbox';
  import type { OfflineSync } from './sync.svelte';

  /**
   * Waiting offline changes and conflicts (#80): a status line while changes
   * wait or sync, and a dialog to decide refused ones (keep mine / keep the
   * server's version, or discard a rejected change).
   */
  interface Props {
    sync: OfflineSync;
  }
  let { sync }: Props = $props();
  let open = $state(false);

  const conflicts = $derived(sync.state.conflicts);
  $effect(() => {
    if (conflicts.length === 0) open = false;
  });

  const what = (c: Conflict) => {
    const p = c.change.payload as { title?: string; body?: string };
    const name = p.title || (p.body ?? '').split('\n')[0] || '…';
    return `${t(c.change.entity === 'todo' ? 'offline.todo' : 'offline.note')}: ${name}`;
  };
</script>

{#if sync.pending > 0 || conflicts.length > 0}
  <div class="sync" role="status">
    {#if sync.pending > 0}
      <span>{sync.syncing ? t('offline.syncing') : t('offline.pending', { n: sync.pending })}</span>
      {#if !sync.syncing}<button class="btn" type="button" onclick={() => sync.replay()}
          >{t('offline.sync_now')}</button
        >{/if}
    {/if}
    {#if conflicts.length > 0}
      <span class="conflict">{t('offline.conflicts', { n: conflicts.length })}</span>
      <button class="btn btn-primary" type="button" onclick={() => (open = true)}
        >{t('offline.resolve')}</button
      >
    {/if}
  </div>
{/if}

{#if open}
  <Modal title={t('offline.resolve_title')} onclose={() => (open = false)}>
    <ul class="conflicts">
      {#each conflicts as c (c.change.key)}
        <li>
          <strong>{what(c)}</strong>
          <p class="muted small">
            {c.kind === 'changed'
              ? t('offline.changed_elsewhere')
              : t('offline.rejected', { message: c.message })}
          </p>
          <div class="actions">
            {#if c.kind === 'changed'}
              <button class="btn btn-primary" type="button" onclick={() => sync.keepMine(c.change.key)}
                >{t('offline.keep_mine')}</button
              >
              <button class="btn" type="button" onclick={() => sync.discard(c.change.key)}
                >{t('offline.keep_server')}</button
              >
            {:else}
              <button class="btn" type="button" onclick={() => sync.discard(c.change.key)}
                >{t('offline.discard')}</button
              >
            {/if}
          </div>
        </li>
      {/each}
    </ul>
  </Modal>
{/if}

<style>
  .sync {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin: 0 0 var(--space-2);
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-small);
    background: color-mix(in srgb, var(--color-warning) 18%, transparent);
  }
  .conflict {
    font-weight: 600;
  }
  .conflicts {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-3);
  }
  .actions {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  p {
    margin: var(--space-1) 0;
  }
  .small {
    font-size: 0.85em;
  }
</style>
