<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type MediaProvider, providerName } from '../domain/media';
  import type { MediaUseCases } from '../usecase/mediaUseCases';

  /** Props of the media server connection settings. */
  interface Props {
    media: MediaUseCases;
  }
  let { media }: Props = $props();

  let provider = $state<MediaProvider>('jellyfin');
  let storedProvider = $state<MediaProvider | null>(null);
  let url = $state('');
  let token = $state('');
  let hasToken = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);
  const keepToken = $derived(hasToken && provider === storedProvider);

  $effect(() => {
    media
      .settings()
      .then((s) => {
        provider = s.provider;
        url = s.url;
        hasToken = s.hasToken;
        storedProvider = s.configured ? s.provider : null;
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await media.saveSettings(provider, url, token);
      hasToken = s.hasToken;
      storedProvider = s.provider;
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
  <label class="field"
    ><span>{t('media.server')}</span><select class="input" bind:value={provider}>
      {#each ['jellyfin', 'plex'] as const as p (p)}<option value={p}>{providerName(p)}</option>{/each}
    </select></label
  >
  <label class="field"
    ><span>{t('media.url', { name: providerName(provider) })}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder={provider === 'plex' ? 'http://192.168.1.2:32400' : 'http://192.168.1.2:8096'}
      required
    /></label
  >
  <label class="field">
    <span
      >{t(provider === 'plex' ? 'media.plex_token' : 'media.jellyfin_key', {
        value: keepToken ? t('common.keep_stored') : '',
      })}</span
    >
    <input
      class="input"
      type="password"
      bind:value={token}
      autocomplete="new-password"
      required={!keepToken}
    />
  </label>
  <p class="muted">{t(provider === 'plex' ? 'media.plex_hint' : 'media.jellyfin_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
