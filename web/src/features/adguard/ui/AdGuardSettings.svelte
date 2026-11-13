<script lang="ts">
  import PlainHttpWarning from '../../../shared/ui/PlainHttpWarning.svelte';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type DnsProvider, dnsProviders, needsUsername, providerName } from '../domain/adguard';
  import type { AdGuardUseCases } from '../usecase/adguardUseCases';

  /** Props of the DNS filter connection settings (AdGuard Home and Pi-hole). */
  interface Props {
    adguard: AdGuardUseCases;
  }
  let { adguard }: Props = $props();

  let provider = $state<DnsProvider>('adguard');
  let url = $state('');
  let username = $state('');
  let password = $state('');
  let hasPassword = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    const p = provider;
    error = message = password = '';
    adguard
      .settings(p)
      .then((s) => ((url = s.url), (username = s.username), (hasPassword = s.hasPassword)))
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await adguard.saveSettings(provider, url, username, password);
      hasPassword = s.hasPassword;
      password = '';
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
    ><span>{t('adguard.provider')}</span><select class="input" bind:value={provider}>
      {#each dnsProviders as p (p)}<option value={p}>{providerName(p)}</option>{/each}
    </select></label
  >
  <label class="field"
    ><span>{t('adguard.url', { name: providerName(provider) })}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder={provider === 'pihole' ? 'http://192.168.1.2' : 'http://192.168.1.2:3000'}
      required
    /></label
  >
  <PlainHttpWarning {url} />
  {#if needsUsername(provider)}
    <label class="field"
      ><span>{t('adguard.username')}</span><input
        class="input"
        bind:value={username}
        autocomplete="off"
        required
      /></label
    >
  {/if}
  <label class="field">
    <span>{t('adguard.password', { value: hasPassword ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={password}
      autocomplete="new-password"
      required={!hasPassword && needsUsername(provider)}
    />
  </label>
  {#if provider === 'pihole'}<p class="muted">{t('adguard.pihole_hint')}</p>{/if}
  <p class="muted">{t('adguard.the_password_is_stored_encrypted_and')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
