<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { emptyRule, type WasteRule } from '../domain/waste';
  import type { WasteUseCases } from '../usecase/wasteUseCases';

  /** Props of the waste calendar settings. */
  interface Props {
    waste: WasteUseCases;
  }
  let { waste }: Props = $props();

  let url = $state('');
  let rules = $state<WasteRule[]>([]);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    waste.settings().then((s) => ((url = s.url), (rules = s.rules)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await waste.saveSettings(url, rules);
      rules = s.rules;
      message = 'Loaded and saved.';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">
  Many municipalities offer the collection dates as an ICS/iCal link (sometimes <code>webcal://</code>). Types
  are recognised automatically (Restmüll, Bio, Papier, Gelber Sack, Glas, …); rules below override label,
  color and icon.
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>ICS link</span><input
      class="input"
      bind:value={url}
      placeholder="https://…/abfuhrkalender.ics"
      required
    /></label
  >
  <fieldset>
    <legend>Rules (optional)</legend>
    {#each rules as r, i (i)}
      <div class="rule">
        <input
          class="input"
          bind:value={r.match}
          placeholder="contains …"
          aria-label="Match text"
          maxlength="40"
        />
        <input class="input" bind:value={r.label} placeholder="Label" aria-label="Label" maxlength="40" />
        <input class="input icon" bind:value={r.icon} placeholder="🗑️" aria-label="Icon" maxlength="8" />
        <input
          type="color"
          value={r.color || '#64748b'}
          oninput={(e) => (r.color = e.currentTarget.value)}
          aria-label="Color"
        />
        <button class="btn btn-icon" type="button" aria-label="Remove rule" onclick={() => rules.splice(i, 1)}
          >✕</button
        >
      </div>
    {/each}
    <button class="btn" type="button" onclick={() => rules.push(emptyRule())} disabled={rules.length >= 20}
      >Add rule</button
    >
  </fieldset>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}>{busy ? 'Loading…' : 'Test & save'}</button>
</form>

<style>
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    margin: var(--space-3) 0;
    display: grid;
    gap: var(--space-1);
  }
  .rule {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  .rule .input {
    flex: 1;
    min-width: 0;
  }
  .rule .icon {
    flex: 0 0 3.5rem;
  }
</style>
