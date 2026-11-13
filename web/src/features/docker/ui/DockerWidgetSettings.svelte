<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseDockerConfig } from '../domain/container';
  import type { DockerUseCases } from '../usecase/dockerUseCases';

  /** Props of the docker widget settings. */
  interface Props {
    docker: DockerUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { docker, config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parseDockerConfig(config));
  let projects = $state<string[]>([]);
  $effect(() => {
    docker
      .containers()
      .then((d) => (projects = [...new Set(d.containers.map((c) => c.project).filter(Boolean))].sort()));
  });
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <label class="field">
    <span>{t('docker.compose_project')}</span>
    <select class="input" bind:value={form.project}>
      <option value="">{t('docker.all_containers')}</option>
      {#each projects as p (p)}<option value={p}>{p}</option>{/each}
    </select>
  </label>
  <label
    ><input type="checkbox" bind:checked={form.problemsOnly} />
    {t('docker.only_show_containers_that_need_attention')}</label
  ><br />
  <label><input type="checkbox" bind:checked={form.stoppedOnly} /> {t('docker.only_stopped')}</label>
  <p><button class="btn btn-primary" type="submit">{t('docker.save')}</button></p>
</form>
