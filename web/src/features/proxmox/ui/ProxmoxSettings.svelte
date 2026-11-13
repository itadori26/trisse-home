<script lang="ts">
  import PlainHttpWarning from '../../../shared/ui/PlainHttpWarning.svelte';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import type { ProxmoxUseCases } from '../usecase/proxmoxUseCases';

  /** Props of the Proxmox connection settings. */
  interface Props {
    proxmox: ProxmoxUseCases;
  }
  let { proxmox }: Props = $props();

  let url = $state('');
  let tokenId = $state('');
  let secret = $state('');
  let fingerprint = $state('');
  let hasSecret = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  /** Shell commands for a read-only token (PVEAuditor). */
  const tokenSnippet = `pveum user add monitor@pve
pveum aclmod / -user monitor@pve -role PVEAuditor
pveum user token add monitor@pve trisse-home --privsep 0`;

  $effect(() => {
    proxmox
      .settings()
      .then(
        (s) => (
          (url = s.url),
          (tokenId = s.tokenId),
          (fingerprint = s.fingerprint),
          (hasSecret = s.hasSecret)
        ),
      )
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await proxmox.saveSettings({ url, tokenId, secret, fingerprint });
      hasSecret = s.hasSecret;
      fingerprint = s.fingerprint;
      secret = '';
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={submit}>
  <p class="muted">{t('proxmox.token_hint')}</p>
  <CopySnippet code={tokenSnippet} label={t('proxmox.token_hint')} />
  <label class="field"
    ><span>{t('proxmox.url')}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="https://192.168.1.10:8006"
      required
    /></label
  >
  <PlainHttpWarning {url} />
  <label class="field"
    ><span>{t('proxmox.token_id')}</span><input
      class="input"
      bind:value={tokenId}
      placeholder="monitor@pve!trisse-home"
      autocomplete="off"
      required
    /></label
  >
  <label class="field">
    <span>{t('proxmox.secret', { value: hasSecret ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={secret}
      autocomplete="new-password"
      required={!hasSecret}
    />
  </label>
  <label class="field"
    ><span>{t('proxmox.fingerprint')}</span><input
      class="input"
      bind:value={fingerprint}
      placeholder="AB:CD:…"
      autocomplete="off"
    /></label
  >
  <p class="muted">{t('proxmox.fingerprint_hint')}</p>
  <p class="muted">{t('proxmox.actions_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
