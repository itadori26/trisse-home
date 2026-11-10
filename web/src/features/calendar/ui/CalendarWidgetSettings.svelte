<script lang="ts">
  import { parseCalendarConfig } from '../domain/calendar';

  /** Props of the calendar widget settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let days = $state(parseCalendarConfig(config).days);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ days });
  }}
>
  <label class="field"
    ><span>Days ahead: {days}</span><input type="range" min="1" max="31" bind:value={days} /></label
  >
  <button class="btn btn-primary" type="submit">Save</button>
</form>
