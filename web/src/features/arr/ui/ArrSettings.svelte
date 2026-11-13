<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    type ArrInstance,
    type ArrInstanceInput,
    appName,
    arrApps,
    defaultPort,
    emptyInstance,
  } from '../domain/arr';
  import type { ArrUseCases } from '../usecase/arrUseCases';

  /** Props of the Sonarr/Radarr/Prowlarr instance management. */
  interface Props {
    arr: ArrUseCases;
  }
  let { arr }: Props = $props();

  let list = $state<ArrInstance[]>([]);
  let form = $state<ArrInstanceInput>(emptyInstance());
  let editing = $state<ArrInstance | null>(null);
  let error = $state('');
  let busy = $state(false);

  async function run(action: () => Promise<void>) {
    error = '';
    busy = true;
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  $effect(() => {
    run(async () => {
      list = await arr.instances();
    });
  });

  function edit(i: ArrInstance) {
    editing = i;
    form = { app: i.app, name: i.name, url: i.url, apiKey: '' };
  }

  function reset() {
    editing = null;
    form = emptyInstance();
  }

  function submit(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      await arr.save(form, editing?.id);
      list = await arr.instances();
      reset();
    });
  }

  function remove(i: ArrInstance) {
    if (!confirm(t('common.confirm_delete', { name: i.name }))) return;
    run(async () => {
      await arr.delete(i.id);
      list = list.filter((x) => x.id !== i.id);
      if (editing?.id === i.id) reset();
    });
  }
</script>

<ul class="list">
  {#each list as i (i.id)}
    <li>
      <span class="name">{i.name} <small class="muted">{appName(i.app)} · {i.url}</small></span>
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('arr.edit', { name: i.name })}
        onclick={() => edit(i)}>✎</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('arr.delete', { name: i.name })}
        onclick={() => remove(i)}>✕</button
      >
    </li>
  {/each}
</ul>
<form onsubmit={submit}>
  <div class="row">
    <label class="field"
      ><span>{t('arr.app')}</span><select class="input" bind:value={form.app}>
        {#each arrApps as a (a)}<option value={a}>{appName(a)}</option>{/each}
      </select></label
    >
    <label class="field"
      ><span>{t('arr.name')}</span><input
        class="input"
        bind:value={form.name}
        maxlength="40"
        placeholder={appName(form.app)}
      /></label
    >
  </div>
  <label class="field"
    ><span>URL</span><input
      class="input"
      type="url"
      bind:value={form.url}
      required
      placeholder="http://192.168.1.2:{defaultPort(form.app)}"
    /></label
  >
  <label class="field"
    ><span>{t('arr.api_key', { value: editing ? t('common.keep_stored') : '' })}</span><input
      class="input"
      type="password"
      bind:value={form.apiKey}
      autocomplete="new-password"
      required={!editing}
    /></label
  >
  <p class="muted">{t('arr.key_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p>
    <button class="btn btn-primary" type="submit" disabled={busy}
      >{busy ? t('common.testing') : editing ? t('common.save') : t('arr.add')}</button
    >
    {#if editing}<button class="btn" type="button" onclick={reset}>{t('common.cancel')}</button>{/if}
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
