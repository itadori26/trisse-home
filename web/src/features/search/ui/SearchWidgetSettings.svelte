<script lang="ts">
  /** Props of the per-widget search settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let newTab = $state(config['newTab'] === true);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ newTab });
  }}
>
  <label><input type="checkbox" bind:checked={newTab} /> Open results in a new tab</label>
  <p class="muted">The search engine and custom shortcuts are configured under Settings → Search.</p>
  <button class="btn btn-primary" type="submit">Save</button>
</form>
