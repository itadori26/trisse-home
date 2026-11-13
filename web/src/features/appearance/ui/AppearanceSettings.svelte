<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Appearance } from '../domain/appearance';
  import type { AppearanceUseCases } from '../usecase/appearanceUseCases';
  import ThemeGallery from './ThemeGallery.svelte';

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
    <span>{t('appearance.greeting_name')}</span>
    <input class="input" bind:value={form.greetingName} maxlength="40" autocomplete="given-name" />
  </label>
  <label class="field" id="settings-launcher">
    <span>{t('appearance.launcher')}</span>
    <select class="input" bind:value={form.launcher}>
      <option value="both">{t('appearance.launcher_both')}</option>
      <option value="services">{t('appearance.launcher_services')}</option>
      <option value="bookmarks">{t('appearance.launcher_bookmarks')}</option>
    </select>
  </label>
  <label class="field">
    <span>{t('appearance.language')}</span>
    <select class="input" bind:value={form.language}>
      <option value="">{t('appearance.language_browser')}</option>
      <option value="en">English</option>
      <option value="de">Deutsch</option>
    </select>
  </label>
  <label class="field">
    <span>{t('appearance.theme')}</span>
    <select class="input" bind:value={form.theme}>
      <option value="system">{t('appearance.system')}</option>
      <option value="light">{t('appearance.light')}</option>
      <option value="dark">{t('appearance.dark')}</option>
    </select>
  </label>
  <label class="field">
    <span>{t('appearance.accent_color')}</span>
    <span class="color">
      <input
        type="color"
        value={form.accentColor || '#3b6ef5'}
        oninput={(e) => (form.accentColor = e.currentTarget.value)}
      />
      <input class="input" bind:value={form.accentColor} placeholder={t('appearance.default')} />
    </span>
  </label>
  <label class="field">
    <span>{t('appearance.background_image_url')}</span>
    <input class="input" type="url" bind:value={form.backgroundImage} placeholder="https://…" />
  </label>
  <label class="field">
    <span>{t('appearance.custom_css_override_variables_such_as')}</span>
    <textarea class="input css" bind:value={form.customCss} rows="6" spellcheck="false"></textarea>
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if saved}<p class="muted">Saved.</p>{/if}
  <button class="btn btn-primary" type="submit">{t('appearance.save_appearance')}</button>
</form>

<ThemeGallery
  {appearance}
  onapply={(a) => {
    form = { ...a };
    saved = false;
    onchange(a);
  }}
/>

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
