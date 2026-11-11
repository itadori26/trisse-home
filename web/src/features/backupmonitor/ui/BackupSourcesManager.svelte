<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { type BackupSource, type CreatedSource, pingExamples } from '../domain/source';
  import type { BackupMonitorUseCases } from '../usecase/backupMonitorUseCases';

  /** Props of the backup source management section. */
  interface Props {
    backupMonitor: BackupMonitorUseCases;
  }
  let { backupMonitor }: Props = $props();

  let sources = $state<BackupSource[]>([]);
  let created = $state<CreatedSource | null>(null);
  let name = $state('');
  let maxAgeHours = $state(26);
  let error = $state('');

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      sources = await backupMonitor.list();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    run(async () => {});
  });
</script>

<p class="muted">
  Each source gets a secret ping URL. Call it from your backup script after every run; the source is reported
  as overdue when no successful run arrives within the maximum age.
</p>
{#if created}
  <div class="created card">
    <p><strong>Ping URL for “{created.source.name}”</strong> — shown only now:</p>
    <pre>{pingExamples(created.pingUrl)}</pre>
    <button class="btn" type="button" onclick={() => (created = null)}>Done</button>
  </div>
{/if}
<ul class="list">
  {#each sources as s (s.id)}
    <li>
      <span class="name">{s.name} <span class="muted">(max. {s.maxAgeHours} h, {s.state})</span></span>
      <button
        class="btn btn-icon"
        type="button"
        onclick={() =>
          confirm(`Create a new ping URL for "${s.name}"? The old one stops working.`) &&
          run(async () => (created = await backupMonitor.rotateToken(s.id)))}>New URL</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label="Delete {s.name}"
        onclick={() => confirm(`Delete "${s.name}"?`) && run(() => backupMonitor.delete(s.id))}>✕</button
      >
    </li>
  {/each}
</ul>
<form
  class="add"
  onsubmit={(e) => {
    e.preventDefault();
    run(async () => {
      created = await backupMonitor.create(name, maxAgeHours);
      name = '';
    });
  }}
>
  <input
    class="input"
    bind:value={name}
    placeholder="e.g. NAS restic"
    aria-label="Name"
    required
    maxlength="60"
  />
  <label class="hours"
    ><input
      class="input"
      type="number"
      min="1"
      max="1440"
      bind:value={maxAgeHours}
      aria-label="Maximum age in hours"
    /> h</label
  >
  <button class="btn" type="submit">Add</button>
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    align-items: center;
  }
  .name {
    flex: 1;
  }
  .add {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  .hours {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .hours .input {
    width: 5rem;
  }
  .created {
    padding: var(--space-3);
    margin-bottom: var(--space-3);
  }
  pre {
    overflow: auto;
    font-size: 0.8rem;
    background: var(--color-surface-2);
    padding: var(--space-2);
    border-radius: var(--radius-small);
  }
</style>
