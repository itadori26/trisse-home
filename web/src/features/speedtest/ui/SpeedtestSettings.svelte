<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { SpeedtestUseCases } from '../usecase/speedtestUseCases';

  /** Props of the Speedtest Tracker connection settings. */
  interface Props {
    speedtest: SpeedtestUseCases;
  }
  let { speedtest }: Props = $props();

  let url = $state('');
  let token = $state('');
  let hasToken = $state(false);
  let minDownMbps = $state(0);
  let minUpMbps = $state(0);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    speedtest
      .settings()
      .then(
        (s) => (
          (url = s.url),
          (hasToken = s.hasToken),
          (minDownMbps = s.minDownMbps),
          (minUpMbps = s.minUpMbps)
        ),
      )
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await speedtest.saveSettings({ url, token, minDownMbps, minUpMbps });
      hasToken = s.hasToken;
      token = '';
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={submit}>
  <p class="muted">{t('speedtest.settings_hint')}</p>
  <label class="field">
    <span>{t('speedtest.url')}</span>
    <input class="input" type="url" bind:value={url} placeholder="http://192.168.1.2:8765" required />
  </label>
  <label class="field">
    <span>{t('speedtest.token', { value: hasToken ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={token}
      autocomplete="new-password"
      required={!hasToken}
    />
  </label>
  <label class="field">
    <span>{t('speedtest.min_down')}</span>
    <input class="input" type="number" min="0" step="1" bind:value={minDownMbps} />
  </label>
  <label class="field">
    <span>{t('speedtest.min_up')}</span>
    <input class="input" type="number" min="0" step="1" bind:value={minUpMbps} />
  </label>
  <p class="muted">{t('speedtest.threshold_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
