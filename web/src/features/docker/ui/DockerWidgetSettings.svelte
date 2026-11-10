<script lang="ts">
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
    <span>Compose project</span>
    <select class="input" bind:value={form.project}>
      <option value="">All containers</option>
      {#each projects as p (p)}<option value={p}>{p}</option>{/each}
    </select>
  </label>
  <label
    ><input type="checkbox" bind:checked={form.problemsOnly} /> Only show containers that need attention</label
  >
  <p><button class="btn btn-primary" type="submit">Save</button></p>
</form>
