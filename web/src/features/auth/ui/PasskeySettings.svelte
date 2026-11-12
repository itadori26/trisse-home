<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import type { Passkey } from '../domain/passkey';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the passkey section in Settings → Account. */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let available = $state(false);
  let passkeys = $state<Passkey[]>([]);
  let name = $state('');
  let codes = $state<string[]>([]);
  let error = $state('');
  let busy = $state(false);

  const load = () =>
    auth
      .passkeys()
      .then((r) => ((available = r.available), (passkeys = r.passkeys)))
      .catch((err) => (error = errorMessage(err)));

  $effect(() => {
    load();
  });

  async function add(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    try {
      codes = await auth.addPasskey(name || t('auth.passkey_default_name'));
      name = '';
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove(p: Passkey) {
    if (!confirm(t('auth.passkey_delete_confirm', { name: p.name }))) return;
    try {
      await auth.deletePasskey(p.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<h3>{t('auth.passkeys')}</h3>
<p class="muted">{t('auth.passkeys_hint')}</p>
{#if !available || !auth.passkeysSupported()}
  <p class="muted" role="note">{t('auth.passkeys_need_https')}</p>
{:else}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if codes.length}
    <div class="codes" role="status">
      <p><strong>{t('auth.passkey_recovery_codes')}</strong></p>
      <ul>
        {#each codes as c (c)}<li><code>{c}</code></li>{/each}
      </ul>
      <button class="btn" type="button" onclick={() => (codes = [])}>{t('auth.i_saved_them')}</button>
    </div>
  {/if}
  <ul class="list">
    {#each passkeys as p (p.id)}
      <li>
        <span
          ><strong>{p.name}</strong>
          <span class="muted small"
            >{t('auth.passkey_last_used', {
              when: p.lastUsedAt ? p.lastUsedAt.toLocaleString(intlLocale()) : t('auth.token_never'),
            })}</span
          ></span
        >
        <button class="btn btn-danger" type="button" onclick={() => remove(p)}
          >{t('auth.passkey_delete')}</button
        >
      </li>
    {:else}
      <li class="muted">{t('auth.no_passkeys')}</li>
    {/each}
  </ul>
  <form class="add" onsubmit={add}>
    <input
      class="input"
      bind:value={name}
      maxlength="60"
      placeholder={t('auth.passkey_name')}
      aria-label={t('auth.passkey_name')}
    />
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('auth.add_passkey')}</button>
  </form>
{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) 0;
  }
  .add {
    display: flex;
    gap: var(--space-1);
  }
  .add .input {
    flex: 1;
  }
  .codes ul {
    columns: 2;
  }
  .small {
    font-size: 0.85em;
  }
</style>
