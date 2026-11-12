<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { Guest } from '../domain/auth';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the guest management section (owner only). */
  interface Props {
    auth: AuthUseCases;
    /** Dashboard tabs that can be shared. */
    tabs: { id: string; name: string }[];
  }
  let { auth, tabs }: Props = $props();

  let guests = $state<Guest[]>([]);
  let error = $state('');
  let name = $state('');
  let password = $state('');
  let confirmation = $state('');
  let shared = $state<string[]>([]);
  let newPassword = $state<Record<string, string>>({});

  const load = () =>
    auth
      .guests()
      .then((l) => (guests = l))
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

  function create(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      await auth.createGuest(name, password, confirmation, shared);
      name = password = confirmation = '';
      shared = [];
    });
  }

  const toggled = (list: string[], id: string, on: boolean) =>
    on ? [...list, id] : list.filter((x) => x !== id);
</script>

<p class="muted">{t('auth.guests_hint')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}

<ul class="guests">
  {#each guests as g (g.id)}
    <li>
      <strong>{g.name}</strong>
      {#if g.disabled}<span class="muted">({t('auth.guest_disabled')})</span>{/if}
      <fieldset>
        <legend class="muted">{t('auth.guest_tabs')}</legend>
        {#each tabs as tab (tab.id)}
          <label
            ><input
              type="checkbox"
              checked={g.tabs.includes(tab.id)}
              onchange={(e) =>
                run(() => auth.updateGuest(g, toggled(g.tabs, tab.id, e.currentTarget.checked), g.disabled))}
            />
            {tab.name}</label
          >
        {/each}
      </fieldset>
      <div class="row">
        <button class="btn" type="button" onclick={() => run(() => auth.updateGuest(g, g.tabs, !g.disabled))}
          >{g.disabled ? t('auth.guest_enable') : t('auth.guest_disable')}</button
        >
        <input
          class="input"
          type="password"
          autocomplete="new-password"
          placeholder={t('auth.guest_new_password')}
          aria-label={t('auth.guest_new_password_for', { name: g.name })}
          bind:value={newPassword[g.id]}
        />
        <button
          class="btn"
          type="button"
          disabled={!newPassword[g.id]}
          onclick={() =>
            run(async () => {
              await auth.updateGuest(g, g.tabs, g.disabled, newPassword[g.id] ?? '');
              newPassword[g.id] = '';
            })}>{t('auth.guest_set_password')}</button
        >
        <button
          class="btn btn-danger"
          type="button"
          onclick={() =>
            confirm(t('auth.guest_delete_confirm', { name: g.name })) && run(() => auth.deleteGuest(g.id))}
          >{t('auth.guest_delete')}</button
        >
      </div>
    </li>
  {:else}
    <li class="muted">{t('auth.no_guests')}</li>
  {/each}
</ul>

<form class="new" onsubmit={create}>
  <h3>{t('auth.new_guest')}</h3>
  <label class="field"
    ><span>{t('auth.guest_name')}</span><input
      class="input"
      bind:value={name}
      autocomplete="off"
      required
    /></label
  >
  <label class="field"
    ><span>{t('auth.password_at_least_12_characters')}</span><input
      class="input"
      type="password"
      autocomplete="new-password"
      bind:value={password}
      required
    /></label
  >
  <label class="field"
    ><span>{t('auth.repeat_password')}</span><input
      class="input"
      type="password"
      autocomplete="new-password"
      bind:value={confirmation}
      required
    /></label
  >
  <fieldset>
    <legend>{t('auth.guest_tabs')}</legend>
    {#each tabs as tab (tab.id)}
      <label
        ><input
          type="checkbox"
          checked={shared.includes(tab.id)}
          onchange={(e) => (shared = toggled(shared, tab.id, e.currentTarget.checked))}
        />
        {tab.name}</label
      >
    {/each}
  </fieldset>
  <button class="btn btn-primary" type="submit">{t('auth.create_guest')}</button>
</form>

<style>
  .guests {
    list-style: none;
    padding: 0;
  }
  .guests li {
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  fieldset {
    border: 0;
    padding: 0;
    margin: var(--space-1) 0;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .row .input {
    flex: 1;
    min-width: 10rem;
  }
  .new {
    margin-top: var(--space-3);
  }
</style>
