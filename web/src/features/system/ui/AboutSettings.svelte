<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { updateHintKey, type VersionInfo } from '../domain/system';
  import type { SystemUseCases } from '../usecase/systemUseCases';

  /** Props of the about section (version and update check). */
  interface Props {
    system: SystemUseCases;
  }
  let { system }: Props = $props();

  let info = $state<VersionInfo | null>(null);
  let error = $state('');

  $effect(() => {
    system
      .version()
      .then((v) => (info = v))
      .catch((err) => (error = errorMessage(err)));
  });

  async function toggle(enabled: boolean) {
    error = '';
    try {
      info = await system.setUpdateCheck(enabled);
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if info}
  <p>{t('system.version_running', { version: info.current })}</p>
  <label>
    <input type="checkbox" checked={info.checkEnabled} onchange={(e) => toggle(e.currentTarget.checked)} />
    {t('system.update_check')}
  </label>
  <p class="muted small">{t('system.update_check_hint')}</p>
  {#if info.installMethod === 'appliance'}
    <p class="muted small">{t('system.update_how_appliance')}</p>
  {/if}
  {#if info.updateAvailable && info.latest}
    {@const hint = updateHintKey(info.installMethod)}
    <p class="update" role="status">
      {t('system.update_available', { version: info.latest.version })}
      <a href={info.latest.url} target="_blank" rel="noopener noreferrer">{t('system.release_notes')}</a>
      {#if hint}<br />{t(hint)}{/if}
    </p>
  {:else if info.checkEnabled && info.latest}
    <p class="muted">{t('system.up_to_date')}</p>
  {/if}
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .update {
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-small);
    background: color-mix(in srgb, var(--color-accent) 18%, transparent);
  }
  .small {
    font-size: 0.85em;
  }
</style>
