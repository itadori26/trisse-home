<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Appearance } from '../domain/appearance';
  import { previewStyle, type ThemeGallery } from '../domain/themes';
  import type { AppearanceUseCases } from '../usecase/appearanceUseCases';

  /** Theme gallery (#111): ready-made themes to apply, import and share. */
  interface Props {
    appearance: AppearanceUseCases;
    /** Called with the preferences after applying or restoring. */
    onapply: (a: Appearance) => void;
  }
  let { appearance, onapply }: Props = $props();

  let gallery = $state<ThemeGallery>({ themes: [], applied: '', canRestore: false });
  let error = $state('');
  let busy = $state(false);
  let picker = $state<HTMLInputElement>();

  const load = async () => (gallery = await appearance.gallery());
  $effect(() => {
    load().catch((err) => (error = errorMessage(err)));
  });

  async function run(action: () => Promise<Appearance | unknown>, apply = false) {
    error = '';
    busy = true;
    try {
      const result = await action();
      if (apply) onapply(result as Appearance);
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function importFile() {
    const file = picker?.files?.[0];
    if (!file) return;
    void run(() => appearance.importTheme(file)).then(() => picker && (picker.value = ''));
  }
</script>

<section aria-labelledby="theme-gallery">
  <h3 id="theme-gallery">{t('appearance.gallery')}</h3>
  <p class="muted small">{t('appearance.gallery_hint')}</p>
  <ul class="grid">
    {#each gallery.themes as theme (theme.id)}
      <li class="card" class:active={gallery.applied === theme.id} style={previewStyle(theme)}>
        <div class="preview" aria-hidden="true">
          <span class="bar"></span>
          <span class="tile"
            ><span class="line"></span><span class="line short"></span><span class="dot"></span></span
          >
        </div>
        <span class="name">{theme.name}</span>
        {#if theme.author}<span class="muted small">{theme.author}</span>{/if}
        <span class="actions">
          <button
            class="btn btn-primary"
            type="button"
            disabled={busy || gallery.applied === theme.id}
            aria-label={t('appearance.apply_theme', { name: theme.name })}
            onclick={() => run(() => appearance.applyTheme(theme.id), true)}
            >{gallery.applied === theme.id ? t('appearance.applied') : t('appearance.apply')}</button
          >
          <a
            class="btn btn-icon"
            href={appearance.exportUrl(theme.id)}
            download
            aria-label={t('appearance.export_theme', { name: theme.name })}>⬇</a
          >
          {#if !theme.builtIn}
            <button
              class="btn btn-icon"
              type="button"
              disabled={busy}
              aria-label={t('appearance.delete_theme', { name: theme.name })}
              onclick={() => run(() => appearance.deleteTheme(theme.id))}>✕</button
            >
          {/if}
        </span>
      </li>
    {/each}
  </ul>
  <div class="row">
    {#if gallery.canRestore}
      <button
        class="btn"
        type="button"
        disabled={busy}
        onclick={() => run(() => appearance.restoreStyle(), true)}>{t('appearance.restore_own')}</button
      >
    {/if}
    <label class="btn"
      >{t('appearance.import_theme')}<input
        class="file"
        bind:this={picker}
        type="file"
        accept="application/json,.json"
        disabled={busy}
        onchange={importFile}
      /></label
    >
  </div>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
  h3 {
    margin: var(--space-4) 0 var(--space-1);
    font-size: 1rem;
  }
  .grid {
    list-style: none;
    padding: 0;
    margin: var(--space-2) 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(9.5rem, 1fr));
    gap: var(--space-2);
  }
  /* The card renders with the theme's own tokens: a live preview. */
  .card {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-2);
    border: 2px solid var(--color-border);
    border-radius: var(--radius);
    background: var(--color-surface);
    color: var(--color-text);
  }
  .card.active {
    border-color: var(--color-accent);
  }
  .card .muted {
    color: var(--color-text-muted);
  }
  .preview {
    display: grid;
    gap: 4px;
    padding: 6px;
    border-radius: var(--radius-small);
    background: var(--color-bg);
    margin-bottom: var(--space-1);
  }
  .bar {
    height: 6px;
    border-radius: 3px;
    background: var(--color-surface-2, var(--color-surface));
  }
  .tile {
    display: grid;
    gap: 3px;
    padding: 6px;
    border-radius: var(--radius-small);
    background: var(--color-surface);
    border: 1px solid var(--color-border);
  }
  .line {
    height: 5px;
    border-radius: 3px;
    background: var(--color-text);
    opacity: 0.8;
  }
  .line.short {
    width: 60%;
    background: var(--color-text-muted);
  }
  .dot {
    width: 30%;
    height: 8px;
    border-radius: 4px;
    background: var(--color-accent);
  }
  .name {
    font-weight: 600;
  }
  .small {
    font-size: 0.8em;
  }
  .actions {
    display: flex;
    gap: 4px;
    margin-top: auto;
    padding-top: var(--space-1);
  }
  .actions .btn-primary {
    flex: 1;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
  }
  .row {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .file {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
  }
  .row label:has(:focus-visible) {
    outline: 2px solid var(--color-accent);
  }
</style>
