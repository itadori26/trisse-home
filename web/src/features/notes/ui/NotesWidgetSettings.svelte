<script lang="ts">
  import { parseNotesConfig } from '../domain/note';

  /** Props of the notes widget settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let scope = $state(parseNotesConfig(config).scope);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ scope });
  }}
>
  <label><input type="radio" bind:group={scope} value="pinned" /> Pinned notes only</label><br />
  <label><input type="radio" bind:group={scope} value="all" /> All notes</label>
  <p><button class="btn btn-primary" type="submit">Save</button></p>
</form>
