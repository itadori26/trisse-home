<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import {
    canAct,
    type ContainerList,
    type DockerActions,
    filterContainers,
    indicator,
    parseDockerConfig,
  } from '../domain/container';
  import { type ImageUpdate, updatesByContainer } from '../domain/imageUpdate';
  import type { DockerUseCases } from '../usecase/dockerUseCases';

  /** Props of the docker widget. */
  interface Props {
    docker: DockerUseCases;
    config: Record<string, unknown>;
  }
  let { docker, config }: Props = $props();

  const POLL_MS = 20_000;
  const cfg = $derived(parseDockerConfig(config));
  let data = $state<ContainerList | null>(null);
  let error = $state('');
  const shown = $derived(data ? filterContainers(data.containers, cfg) : []);
  let actions = $state<DockerActions | null>(null);
  let busy = $state('');
  let logs = $state<{ name: string; lines: string[] | null; error: string } | null>(null);
  let updates = $state(new Map<string, ImageUpdate>());

  // Image update hints (#107); silently absent while disabled.
  $effect(() => {
    const load = () =>
      docker
        .updates()
        .then((l) => (updates = updatesByContainer(l ?? [])))
        .catch(() => {});
    load();
    const timer = setInterval(load, 10 * 60_000);
    return () => clearInterval(timer);
  });

  $effect(() => {
    docker
      .actions()
      .then((a) => (actions = a))
      .catch(() => {});
  });

  async function restart(name: string) {
    if (!confirm(t('docker.confirm_restart', { name }))) return;
    busy = name;
    try {
      await docker.restart(name);
      data = await docker.containers();
      error = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }

  async function showLogs(name: string) {
    logs = { name, lines: null, error: '' };
    try {
      const lines = await docker.logs(name);
      if (logs?.name === name) logs = { name, lines, error: '' };
    } catch (err) {
      if (logs?.name === name) logs = { name, lines: [], error: errorMessage(err) };
    }
  }

  $effect(() => {
    const load = () =>
      docker
        .containers()
        .then((d) => ((data = d), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });
</script>

{#if error}<p class="error">{error}</p>{/if}
{#if data}
  <p class="summary">
    <span><strong>{data.running}</strong> {t('docker.running')}</span>
    <span class:alert={data.stopped > 0}><strong>{data.stopped}</strong> {t('docker.stopped')}</span>
    {#if data.unhealthy}<span class="alert"><strong>{data.unhealthy}</strong> {t('docker.unhealthy')}</span
      >{/if}
  </p>
  {#if data.error}<p class="error small">{data.error}</p>{/if}
  <ul class="list">
    {#each shown as c (c.id)}
      {@const u = updates.get(c.name)}
      <li title="{c.image}\n{c.status}">
        <span class="dot {indicator(c)}" aria-label={c.state}></span>
        <span class="name">{c.name}</span>
        {#if u}
          <a
            class="update"
            href={u.releaseUrl || undefined}
            target="_blank"
            rel="noopener noreferrer"
            title={u.newerTag ? t('docker.newer_tag', { tag: u.newerTag }) : t('docker.update_available')}
            aria-label={t('docker.update_available_for', { name: c.name })}>⬆</a
          >
        {/if}
        <span class="muted small status">{c.status}</span>
        {#if canAct(actions, c.name)}
          <button
            class="btn btn-icon act"
            type="button"
            title={t('docker.show_logs')}
            aria-label={t('docker.logs_of', { name: c.name })}
            onclick={() => showLogs(c.name)}>≡</button
          >
          <button
            class="btn btn-icon act"
            type="button"
            title={t('docker.restart')}
            aria-label={t('docker.restart_2', { name: c.name })}
            disabled={busy === c.name}
            onclick={() => restart(c.name)}>↻</button
          >
        {/if}
      </li>
    {:else}
      <li class="muted">{cfg.problemsOnly ? t('docker.all_healthy') : t('docker.no_containers')}</li>
    {/each}
  </ul>
{/if}

{#if logs}
  <Modal title={t('docker.logs', { name: logs.name })} onclose={() => (logs = null)}>
    {#if logs.error}<p class="error">{logs.error}</p>{/if}
    {#if logs.lines === null}
      <p class="muted">{t('docker.loading')}</p>
    {:else}
      <pre class="logs">{logs.lines.join('\n') || t('docker.no_output')}</pre>
    {/if}
  </Modal>
{/if}

<style>
  .update {
    color: var(--color-warning);
    text-decoration: none;
    font-size: 0.85em;
  }
  .act {
    padding: 0 4px;
    line-height: 1.2;
    font-size: 0.9em;
  }
  .logs {
    max-height: 60vh;
    max-width: min(90vw, 900px);
    overflow: auto;
    font-size: 0.75rem;
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--color-surface-2);
    padding: var(--space-2);
    border-radius: var(--radius-small);
  }
  .summary {
    display: flex;
    gap: var(--space-3);
    margin: 0 0 var(--space-2);
    font-size: 0.85em;
  }
  .alert {
    color: var(--color-danger);
  }
  .small {
    font-size: 0.75em;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 2px 0;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .status {
    white-space: nowrap;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex: none;
  }
  .ok {
    background: var(--color-success);
  }
  .warning {
    background: var(--color-warning);
  }
  .error.dot {
    background: var(--color-danger);
  }
  .off {
    background: var(--color-text-muted);
  }
</style>
