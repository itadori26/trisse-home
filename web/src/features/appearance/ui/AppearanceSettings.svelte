<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Appearance } from '../domain/appearance';
  import type { AppearanceUseCases } from '../usecase/appearanceUseCases';

  /** Props of the appearance settings section. */
  interface Props {
    appearance: AppearanceUseCases;
    current: Appearance;
    /** Called with the saved preferences so the app can apply them. */
    onchange: (a: Appearance) => void;
  }
  let { appearance, current, onchange }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<Appearance>({ ...current });
  let error = $state('');
  let saved = $state(false);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    saved = false;
    try {
      const result = await appearance.save(form);
      onchange(result);
      saved = true;
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field">
    <span>Theme</span>
    <select class="input" bind:value={form.theme}>
      <option value="system">System</option>
      <option value="light">Light</option>
      <option value="dark">Dark</option>
    </select>
  </label>
  <label class="field">
    <span>Accent color</span>
    <span class="color">
      <input
        type="color"
        value={form.accentColor || '#3b6ef5'}
        oninput={(e) => (form.accentColor = e.currentTarget.value)}
      />
      <input class="input" bind:value={form.accentColor} placeholder="default" />
    </span>
  </label>
  <label class="field">
    <span>Background image URL</span>
    <input class="input" type="url" bind:value={form.backgroundImage} placeholder="https://…" />
  </label>
  <label class="field">
    <span>Custom CSS (override variables such as --radius or --color-bg)</span>
    <textarea class="input css" bind:value={form.customCss} rows="6" spellcheck="false"></textarea>
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if saved}<p class="muted">Saved.</p>{/if}
  <button class="btn btn-primary" type="submit">Save appearance</button>
</form>

<style>
  .color {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  .css {
    font-family: var(--font-mono);
    font-size: 0.85rem;
  }
</style>
