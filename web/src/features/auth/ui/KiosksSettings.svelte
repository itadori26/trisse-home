<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { copyToClipboard } from '../../../shared/ui/clipboard';
  import { qrDataUrl } from '../../../shared/ui/qr';
  import { type KioskDisplay, type KioskInput, kioskLink } from '../domain/auth';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Kiosk displays (#50, owner only): wall tablets that show some tabs. */
  interface Props {
    auth: AuthUseCases;
    tabs: { id: string; name: string }[];
  }
  let { auth, tabs }: Props = $props();

  const empty = (): KioskInput => ({
    name: '',
    tabs: [],
    disabled: false,
    rotateSeconds: 0,
    nightFrom: '',
    nightTo: '',
  });
  let kiosks = $state<KioskDisplay[]>([]);
  let form = $state<KioskInput>(empty());
  let error = $state('');
  /** The link of a just created or renewed display; shown once. */
  let link = $state<{ name: string; url: string } | null>(null);
  let copied = $state(false);

  const load = () =>
    auth
      .kiosks()
      .then((l) => (kiosks = l))
      .catch((err) => (error = errorMessage(err)));
  $effect(() => {
    load();
  });

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const show = (name: string, token: string) => {
    link = { name, url: kioskLink(location.origin, token) };
    copied = false;
  };
  const toggled = (list: string[], id: string, on: boolean) =>
    on ? [...list, id] : list.filter((x) => x !== id);
  const input = (k: KioskDisplay): KioskInput => ({
    name: k.name,
    tabs: k.tabs,
    disabled: k.disabled,
    rotateSeconds: k.rotateSeconds,
    nightFrom: k.nightFrom,
    nightTo: k.nightTo,
  });
  const ROTATIONS = [0, 15, 30, 60, 120, 300];
</script>

<p class="muted">{t('auth.kiosk_hint')}</p>
{#if link}
  <div class="link" role="status">
    <p><strong>{t('auth.kiosk_link_for', { name: link.name })}</strong> {t('auth.kiosk_link_once')}</p>
    <img src={qrDataUrl(link.url, 4)} alt={t('auth.kiosk_qr')} width="164" height="164" />
    <code>{link.url}</code>
    <p>
      <button
        class="btn"
        type="button"
        onclick={async () => (copied = await copyToClipboard(link?.url ?? ''))}
        >{copied ? t('auth.kiosk_copied') : t('auth.kiosk_copy')}</button
      >
      <button class="btn" type="button" onclick={() => (link = null)}>{t('auth.kiosk_done')}</button>
    </p>
  </div>
{/if}

<ul class="list">
  {#each kiosks as k (k.id)}
    <li>
      <strong>{k.name}</strong>
      {#if k.disabled}<span class="muted">({t('auth.guest_disabled')})</span>{/if}
      <span class="muted small"
        >{k.lastUsedAt
          ? t('auth.kiosk_last_used', { time: k.lastUsedAt.toLocaleString(intlLocale()) })
          : t('auth.kiosk_never_used')}</span
      >
      <fieldset>
        <legend class="muted">{t('auth.guest_tabs')}</legend>
        {#each tabs as tab (tab.id)}
          <label
            ><input
              type="checkbox"
              checked={k.tabs.includes(tab.id)}
              onchange={(e) =>
                run(() =>
                  auth.updateKiosk(k.id, {
                    ...input(k),
                    tabs: toggled(k.tabs, tab.id, e.currentTarget.checked),
                  }),
                )}
            />
            {tab.name}</label
          >
        {/each}
      </fieldset>
      <p class="muted small">
        {k.rotateSeconds ? t('auth.kiosk_rotates', { n: k.rotateSeconds }) : t('auth.kiosk_no_rotation')}
        {#if k.nightFrom}· {t('auth.kiosk_night_summary', { from: k.nightFrom, to: k.nightTo })}{/if}
      </p>
      <p class="row">
        <button
          class="btn"
          type="button"
          onclick={() => run(async () => show(k.name, await auth.renewKioskLink(k.id)))}
          >{t('auth.kiosk_new_link')}</button
        >
        <button
          class="btn"
          type="button"
          onclick={() => run(() => auth.updateKiosk(k.id, { ...input(k), disabled: !k.disabled }))}
          >{k.disabled ? t('auth.guest_enable') : t('auth.guest_disable')}</button
        >
        <button
          class="btn"
          type="button"
          onclick={() =>
            confirm(t('auth.kiosk_delete_confirm', { name: k.name })) && run(() => auth.deleteKiosk(k.id))}
          >{t('auth.guest_delete')}</button
        >
      </p>
    </li>
  {:else}
    <li class="muted">{t('auth.kiosk_none')}</li>
  {/each}
</ul>

<form
  onsubmit={(e) => {
    e.preventDefault();
    void run(async () => {
      const { kiosk, token } = await auth.createKiosk(form);
      show(kiosk.name, token);
      form = empty();
    });
  }}
>
  <h3>{t('auth.kiosk_new')}</h3>
  <label class="field"
    ><span>{t('auth.kiosk_name')}</span><input
      class="input"
      bind:value={form.name}
      placeholder="flur-tablet"
      required
    /></label
  >
  <fieldset>
    <legend>{t('auth.guest_tabs')}</legend>
    {#each tabs as tab (tab.id)}
      <label
        ><input
          type="checkbox"
          checked={form.tabs.includes(tab.id)}
          onchange={(e) => (form.tabs = toggled(form.tabs, tab.id, e.currentTarget.checked))}
        />
        {tab.name}</label
      >
    {/each}
  </fieldset>
  <label class="field"
    ><span>{t('auth.kiosk_rotation')}</span>
    <select class="input" bind:value={form.rotateSeconds}>
      {#each ROTATIONS as s (s)}<option value={s}
          >{s ? t('auth.kiosk_rotates', { n: s }) : t('auth.kiosk_no_rotation')}</option
        >{/each}
    </select>
  </label>
  <div class="row">
    <label class="field"
      ><span>{t('auth.kiosk_night_from')}</span><input
        class="input"
        type="time"
        bind:value={form.nightFrom}
      /></label
    >
    <label class="field"
      ><span>{t('auth.kiosk_night_to')}</span><input
        class="input"
        type="time"
        bind:value={form.nightTo}
      /></label
    >
  </div>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={!form.name || form.tabs.length === 0}
    >{t('auth.kiosk_create')}</button
  >
</form>

<style>
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-3);
  }
  fieldset {
    border: none;
    padding: 0;
    margin: var(--space-1) 0;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .link {
    padding: var(--space-3);
    border: 2px solid var(--color-accent);
    border-radius: var(--radius);
    margin-bottom: var(--space-3);
  }
  .link code {
    display: block;
    word-break: break-all;
    margin: var(--space-2) 0;
  }
  h3 {
    font-size: 1rem;
    margin: var(--space-4) 0 var(--space-2);
  }
  .small {
    font-size: 0.85em;
  }
</style>
