<script lang="ts">
  import { type ClockConfig, isValidTimeZone, parseClockConfig } from '../domain/clock';

  /** Props of the clock settings form. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<ClockConfig>(parseClockConfig(config));
  let error = $state('');

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    if (!isValidTimeZone(form.timeZone.trim())) {
      error = 'Unknown time zone. Use an IANA name such as Europe/Berlin.';
      return;
    }
    await onsave({ ...form, timeZone: form.timeZone.trim() });
  }
</script>

<form onsubmit={submit}>
  <label class="field">
    <span>Label</span>
    <input class="input" bind:value={form.label} maxlength="40" />
  </label>
  <label class="field">
    <span>Time zone (empty = this device)</span>
    <input class="input" bind:value={form.timeZone} placeholder="Europe/Berlin" />
  </label>
  <label><input type="checkbox" bind:checked={form.hour12} /> 12-hour clock</label><br />
  <label><input type="checkbox" bind:checked={form.showSeconds} /> Show seconds</label><br />
  <label><input type="checkbox" bind:checked={form.showDate} /> Show date</label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p><button class="btn btn-primary" type="submit">Save</button></p>
</form>
