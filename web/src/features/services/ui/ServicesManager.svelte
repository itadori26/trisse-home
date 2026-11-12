<script lang="ts">
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import { type WolSetup, wolEnvSnippet } from '../domain/wolSnippet';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { emptyServiceInput, type Service, type ServiceInput, statusText } from '../domain/service';
  import type { ServicesUseCases } from '../usecase/servicesUseCases';
  import ServiceDiscoverySettings from './ServiceDiscoverySettings.svelte';
  import ServiceIcon from './ServiceIcon.svelte';

  /** Props of the services management section in the settings. */
  interface Props {
    services: ServicesUseCases;
  }
  let { services }: Props = $props();

  let list = $state<Service[]>([]);
  let form = $state<ServiceInput>(emptyServiceInput());
  let editingId = $state<string | null>(null);
  let error = $state('');

  async function run(action: () => Promise<void>) {
    error = '';
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function edit(s: Service) {
    editingId = s.id;
    form = {
      name: s.name,
      url: s.url,
      checkUrl: s.checkUrl === s.url ? '' : s.checkUrl,
      icon: s.icon,
      description: s.description,
      skipTlsVerify: s.skipTlsVerify,
      mac: s.mac,
    };
  }

  function reset() {
    editingId = null;
    form = emptyServiceInput();
  }

  function submit(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      await services.save(form, editingId ?? undefined);
      list = await services.listAll();
      reset();
    });
  }

  function remove(s: Service) {
    if (!confirm(t('common.confirm_delete', { name: s.name }))) return;
    run(async () => {
      await services.delete(s.id);
      list = list.filter((x) => x.id !== s.id);
    });
  }

  run(async () => {
    list = await services.listAll();
  });

  let wolSetup = $state<WolSetup>('relay');
  let wolBroadcast = $state('');
  let wolDockerActions = $state(false);
  const wolSnippet = $derived(
    wolEnvSnippet({ setup: wolSetup, broadcast: wolBroadcast, dockerActions: wolDockerActions }),
  );
</script>

<ul class="list">
  {#each list as s, i (s.id)}
    <li>
      <ServiceIcon icon={s.icon} name={s.name} />
      <span class="name" title={statusText(s)}
        >{s.name}
        {#if s.hidden}
          <small class="muted">{t('services.hidden_hint', { container: s.container })}</small>
        {:else if s.container}
          <small class="muted">{t('services.discovered', { container: s.container })}</small>
        {/if}
      </span>
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('services.move_up')}
        disabled={i === 0}
        onclick={() =>
          run(async () => {
            list = await services.move(list, s.id, -1);
          })}>↑</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('services.move_down')}
        disabled={i === list.length - 1}
        onclick={() =>
          run(async () => {
            list = await services.move(list, s.id, 1);
          })}>↓</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('services.edit', { name: s.name })}
        onclick={() => edit(s)}>✎</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('services.delete', { name: s.name })}
        onclick={() => remove(s)}>✕</button
      >
    </li>
  {/each}
</ul>

<ServiceDiscoverySettings
  {services}
  onsynced={() =>
    run(async () => {
      list = await services.listAll();
    })}
/>

<form onsubmit={submit}>
  <h4>{editingId ? t('services.edit_service') : t('services.add_service')}</h4>
  <div class="row">
    <label class="field"
      ><span>{t('services.name')}</span><input
        class="input"
        bind:value={form.name}
        required
        maxlength="60"
      /></label
    >
    <label class="field"
      ><span>{t('services.icon_emoji_or_image_url')}</span><input
        class="input"
        bind:value={form.icon}
      /></label
    >
  </div>
  <label class="field"
    ><span>URL</span><input
      class="input"
      type="url"
      bind:value={form.url}
      required
      placeholder="https://cloud.example.com"
    /></label
  >
  <label class="field"
    ><span>{t('services.health_check_url_optional')}</span><input
      class="input"
      type="url"
      bind:value={form.checkUrl}
      placeholder={t('services.defaults_to_the_url')}
    /></label
  >
  <label class="field"
    ><span>{t('services.description')}</span><input
      class="input"
      bind:value={form.description}
      maxlength="200"
    /></label
  >
  <label class="field"
    ><span>{t('services.mac_address_for_wake_on_lan')}</span><input
      class="input"
      bind:value={form.mac}
      placeholder="aa:bb:cc:dd:ee:ff"
      autocomplete="off"
    /></label
  >
  {#if form.mac.trim() !== ''}
    <details>
      <summary>{t('services.wol_setup')}</summary>
      <label class="field"
        ><span>{t('services.wol_how')}</span><select class="input" bind:value={wolSetup}>
          <option value="relay">{t('services.wol_relay')}</option>
          <option value="host-network">{t('services.wol_host_network')}</option>
        </select></label
      >
      <label class="field"
        ><span>{t('services.wol_broadcast')}</span><input
          class="input"
          bind:value={wolBroadcast}
          placeholder="192.168.178.255:9"
          autocomplete="off"
        /></label
      >
      <label
        ><input type="checkbox" bind:checked={wolDockerActions} /> {t('services.wol_docker_actions')}</label
      >
      {#if wolSnippet}
        <p class="muted">{t('services.wol_snippet_hint')}</p>
        <CopySnippet code={wolSnippet} label={t('services.wol_setup')} />
      {:else}
        <p class="error">{t('services.wol_broadcast_invalid')}</p>
      {/if}
      <p class="muted">
        <a
          href="https://github.com/itadori26/trisse-home/blob/main/docs/security-model.md#wake-on-lan-relay"
          target="_blank"
          rel="noopener noreferrer">{t('common.why_extra_step')}</a
        >
      </p>
    </details>
  {/if}
  <label
    ><input type="checkbox" bind:checked={form.skipTlsVerify} />
    {t('services.accept_self_signed_certificates_for_the')}</label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p>
    <button class="btn btn-primary" type="submit">{editingId ? t('common.save') : t('common.add')}</button>
    {#if editingId}<button class="btn" type="button" onclick={reset}>{t('services.cancel')}</button>{/if}
  </p>
</form>

<style>
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .list li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .name {
    flex: 1;
    display: grid;
  }
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }
</style>
