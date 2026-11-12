<script lang="ts">
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { type Discovery, labelExample, MAX_INTERVAL, MIN_INTERVAL } from '../domain/discovery';
  import type { ServicesUseCases } from '../usecase/servicesUseCases';

  /** Props of the Docker autodiscovery settings (#78). */
  interface Props {
    services: ServicesUseCases;
    /** Called after a sync so the tile list can reload. */
    onsynced: () => void;
  }
  let { services, onsynced }: Props = $props();

  let current = $state<Discovery | null>(null);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  async function run(action: () => Promise<void>) {
    error = message = '';
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
    services
      .discovery()
      .then((d) => (current = d))
      .catch((err) => (error = errorMessage(err)));
  });

  function submit(e: SubmitEvent) {
    e.preventDefault();
    const s = current;
    if (!s) return;
    run(async () => {
      current = await services.saveDiscovery({
        enabled: s.enabled,
        intervalMinutes: s.intervalMinutes,
        rule: s.rule,
      });
      message = t('common.saved');
    });
  }

  function syncNow() {
    run(async () => {
      current = await services.syncDiscovery();
      onsynced();
    });
  }
</script>

<details>
  <summary>{t('services.discovery.title')}</summary>
  {#if current && !current.available}
    <p class="muted">{t('services.discovery.unavailable')}</p>
  {:else if current}
    <form onsubmit={submit}>
      <p class="muted">{t('services.discovery.hint')}</p>
      <CopySnippet code={labelExample} label={t('services.discovery.title')} />
      <label class="check"
        ><input type="checkbox" bind:checked={current.enabled} /> {t('services.discovery.enabled')}</label
      >
      <div class="row">
        <label class="field"
          ><span>{t('services.discovery.interval')}</span><input
            class="input"
            type="number"
            min={MIN_INTERVAL}
            max={MAX_INTERVAL}
            bind:value={current.intervalMinutes}
          /></label
        >
        <label class="field"
          ><span>{t('services.discovery.rule')}</span><select class="input" bind:value={current.rule}>
            <option value="labels">{t('services.discovery.rule_labels')}</option>
            <option value="manual">{t('services.discovery.rule_manual')}</option>
          </select></label
        >
      </div>
      {#if current.lastSync}
        <p class="muted" role="status">
          {#if current.lastSync.error}
            {t('services.discovery.last_error', {
              time: current.lastSync.at.toLocaleString(intlLocale()),
              error: current.lastSync.error,
            })}
          {:else}
            {t('services.discovery.last_result', {
              time: current.lastSync.at.toLocaleString(intlLocale()),
              created: current.lastSync.created,
              updated: current.lastSync.updated,
              hidden: current.lastSync.hidden,
            })}
            {#if current.lastSync.skipped > 0}
              {t('services.discovery.skipped', { count: current.lastSync.skipped })}
            {/if}
          {/if}
        </p>
      {/if}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      {#if message}<p class="muted" role="status">{message}</p>{/if}
      <p>
        <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
        <button class="btn" type="button" disabled={busy} onclick={syncNow}
          >{t('services.discovery.sync_now')}</button
        >
      </p>
    </form>
  {/if}
  {#if !current && error}<p class="error" role="alert">{error}</p>{/if}
</details>

<style>
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }
  .check {
    display: block;
    margin-block: var(--space-2);
  }
</style>
