<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { ImageUpdate, RegistryInput } from '../domain/imageUpdate';
  import type { DockerUseCases } from '../usecase/dockerUseCases';

  /** Props of the image update settings and overview (#107). */
  interface Props {
    docker: DockerUseCases;
  }
  let { docker }: Props = $props();

  let enabled = $state(false);
  let semverTags = $state(false);
  let registries = $state<(RegistryInput & { hasPassword: boolean })[]>([]);
  let images = $state<ImageUpdate[] | null>(null);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  function loadImages() {
    docker
      .updates()
      .then((l) => (images = l))
      .catch((err) => (error = errorMessage(err)));
  }

  $effect(() => {
    docker
      .updateSettings()
      .then((s) => {
        enabled = s.enabled;
        semverTags = s.semverTags;
        registries = s.registries.map((r) => ({ ...r, password: '' }));
        if (s.enabled) loadImages();
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await docker.saveUpdateSettings(enabled, semverTags, registries);
      registries = s.registries.map((r) => ({ ...r, password: '' }));
      message = t('common.saved');
      if (s.enabled) loadImages();
      else images = null;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={submit}>
  <p class="muted">{t('docker_updates.hint')}</p>
  <label class="check"><input type="checkbox" bind:checked={enabled} /> {t('docker_updates.enabled')}</label>
  <label class="check"><input type="checkbox" bind:checked={semverTags} /> {t('docker_updates.semver')}</label
  >
  <fieldset>
    <legend>{t('docker_updates.registries')}</legend>
    {#each registries as r, i (i)}
      <div class="registry">
        <input
          class="input"
          bind:value={r.host}
          placeholder="ghcr.io"
          aria-label={t('docker_updates.host')}
        />
        <input
          class="input"
          bind:value={r.username}
          autocomplete="off"
          aria-label={t('docker_updates.username')}
        />
        <input
          class="input"
          type="password"
          bind:value={r.password}
          autocomplete="new-password"
          aria-label={t('docker_updates.password', { value: r.hasPassword ? t('common.keep_stored') : '' })}
          placeholder={r.hasPassword ? '••••••' : ''}
        />
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('docker_updates.remove_registry', { host: r.host })}
          onclick={() => (registries = registries.filter((_, j) => j !== i))}>×</button
        >
      </div>
    {/each}
    <button
      class="btn"
      type="button"
      onclick={() =>
        (registries = [...registries, { host: '', username: '', password: '', hasPassword: false }])}
      >{t('docker_updates.add_registry')}</button
    >
  </fieldset>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
</form>

{#if images}
  <h3>{t('docker_updates.images')}</h3>
  <ul class="images">
    {#each images as img (img.ref)}
      <li class="status-{img.status}">
        <code>{img.ref}</code>
        <span class="muted">{img.containers.join(', ')}</span>
        <span>
          {t(`docker_updates.status.${img.status}`)}{img.newerTag
            ? ` · ${t('docker.newer_tag', { tag: img.newerTag })}`
            : ''}
        </span>
        {#if img.releaseUrl}<a href={img.releaseUrl} target="_blank" rel="noopener noreferrer"
            >{t('docker_updates.release_notes')}</a
          >{/if}
        {#if img.error}<span class="error">{img.error}</span>{/if}
      </li>
    {:else}
      <li class="muted">{t('docker.no_containers')}</li>
    {/each}
  </ul>
  {#if images.every((i) => i.status === 'unknown')}<p class="muted">{t('docker_updates.not_checked')}</p>{/if}
{/if}

<style>
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
    margin-bottom: var(--space-1);
  }
  fieldset {
    border: 1px solid var(--color-border, rgba(127, 127, 127, 0.3));
    border-radius: var(--radius-small);
    margin: var(--space-2) 0;
  }
  .registry {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr auto;
    gap: var(--space-1);
    margin-bottom: var(--space-1);
  }
  .images {
    list-style: none;
    padding: 0;
  }
  .images li {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    padding: 2px 0;
  }
  .images .status-update_available {
    color: var(--color-warning);
  }
  @media (max-width: 600px) {
    .registry {
      grid-template-columns: 1fr;
    }
  }
</style>
