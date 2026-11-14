<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { changing, type DeviceState } from '../domain/device';
  import type { DeviceUseCases } from '../usecase/deviceUseCases';

  /** Choice between server only and server + display (#140), with progress. */
  interface Props {
    device: DeviceUseCases;
    /** Called once the choice was sent (the setup step closes then). */
    onchosen?: () => void;
  }
  let { device, onchosen }: Props = $props();

  let current = $state<DeviceState | null>(null);
  let choice = $state<'headless' | 'kiosk'>('headless');
  let error = $state('');
  let busy = $state(false);

  const load = () =>
    device
      .state()
      .then((s) => {
        current = s;
        if (s.requested) choice = s.requested;
      })
      .catch((err) => (error = errorMessage(err)));
  $effect(() => {
    load();
  });
  // While the system installs or removes the display, look every 3 s.
  $effect(() => {
    if (!current || !changing(current)) return;
    const id = setTimeout(load, 3000);
    return () => clearTimeout(id);
  });

  async function apply() {
    error = '';
    busy = true;
    try {
      current = await device.changeMode(choice);
      onchosen?.();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if current}
  <p class="muted" role="status">
    {current.display === 'connected'
      ? t('device.display_connected')
      : current.display === 'disconnected'
        ? t('device.display_disconnected')
        : t('device.display_unknown')}
  </p>
  <fieldset>
    <legend class="sr-only">{t('device.question')}</legend>
    <label class="option"
      ><input type="radio" name="display-mode" value="headless" bind:group={choice} />
      <span
        ><strong>{t('device.headless')}</strong><br /><span class="muted">{t('device.headless_hint')}</span
        ></span
      ></label
    >
    <label class="option"
      ><input type="radio" name="display-mode" value="kiosk" bind:group={choice} />
      <span
        ><strong>{t('device.kiosk')}</strong><br /><span class="muted">{t('device.kiosk_hint')}</span></span
      ></label
    >
  </fieldset>
  {#if changing(current)}
    <p role="status">{current.phase === 'installing' ? t('device.installing') : t('device.pending')}</p>
  {:else if current.phase === 'failed'}
    <p class="error" role="alert">{t('device.failed', { error: current.error })}</p>
  {:else if current.phase === 'active'}
    <p class="muted">{t(current.requested === 'kiosk' ? 'device.active_kiosk' : 'device.active_headless')}</p>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <button class="btn btn-primary" type="button" disabled={busy || changing(current)} onclick={apply}
    >{t('device.apply')}</button
  >
{:else if error}
  <p class="error" role="alert">{error}</p>
{/if}

<style>
  fieldset {
    border: none;
    padding: 0;
    margin: var(--space-2) 0;
    display: grid;
    gap: var(--space-2);
  }
  .option {
    display: flex;
    gap: var(--space-2);
    align-items: flex-start;
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
</style>
