<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { emptyServiceInput, type Service, type ServiceInput, statusText } from '../domain/service';
  import type { ServicesUseCases } from '../usecase/servicesUseCases';
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
      list = await services.list();
      reset();
    });
  }

  function remove(s: Service) {
    if (!confirm(`Delete "${s.name}"?`)) return;
    run(async () => {
      await services.delete(s.id);
      list = list.filter((x) => x.id !== s.id);
    });
  }

  run(async () => {
    list = await services.list();
  });
</script>

<ul class="list">
  {#each list as s, i (s.id)}
    <li>
      <ServiceIcon icon={s.icon} name={s.name} />
      <span class="name" title={statusText(s)}>{s.name}</span>
      <button
        class="btn btn-icon"
        type="button"
        aria-label="Move up"
        disabled={i === 0}
        onclick={() =>
          run(async () => {
            list = await services.move(list, s.id, -1);
          })}>↑</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label="Move down"
        disabled={i === list.length - 1}
        onclick={() =>
          run(async () => {
            list = await services.move(list, s.id, 1);
          })}>↓</button
      >
      <button class="btn btn-icon" type="button" aria-label="Edit {s.name}" onclick={() => edit(s)}>✎</button>
      <button class="btn btn-icon" type="button" aria-label="Delete {s.name}" onclick={() => remove(s)}
        >✕</button
      >
    </li>
  {/each}
</ul>

<form onsubmit={submit}>
  <h4>{editingId ? 'Edit service' : 'Add service'}</h4>
  <div class="row">
    <label class="field"
      ><span>Name</span><input class="input" bind:value={form.name} required maxlength="60" /></label
    >
    <label class="field"
      ><span>Icon (emoji or image URL)</span><input class="input" bind:value={form.icon} /></label
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
    ><span>Health check URL (optional)</span><input
      class="input"
      type="url"
      bind:value={form.checkUrl}
      placeholder="defaults to the URL"
    /></label
  >
  <label class="field"
    ><span>Description</span><input class="input" bind:value={form.description} maxlength="200" /></label
  >
  <label
    ><input type="checkbox" bind:checked={form.skipTlsVerify} /> Accept self-signed certificates for the health
    check</label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p>
    <button class="btn btn-primary" type="submit">{editingId ? 'Save' : 'Add'}</button>
    {#if editingId}<button class="btn" type="button" onclick={reset}>Cancel</button>{/if}
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
  }
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }
</style>
