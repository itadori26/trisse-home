<script lang="ts">
  import { t } from '../../../shared/i18n';
  /** Props of the per-widget search settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  const stored = (key: string, fallback: string) => {
    const p = config['quickPrefixes'];
    const v = p && typeof p === 'object' ? (p as Record<string, unknown>)[key] : undefined;
    return typeof v === 'string' ? v : fallback;
  };
  // svelte-ignore state_referenced_locally
  let newTab = $state(config['newTab'] === true);
  let todo = $state(stored('todo', 't'));
  let note = $state(stored('note', 'n'));
  let bookmark = $state(stored('bookmark', 'b'));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ newTab, quickPrefixes: { todo: todo.trim(), note: note.trim(), bookmark: bookmark.trim() } });
  }}
>
  <label><input type="checkbox" bind:checked={newTab} /> {t('search.open_results_in_a_new_tab')}</label>
  <p class="muted">{t('search.the_search_engine_and_custom_shortcuts')}</p>
  <fieldset>
    <legend>{t('quick.prefixes')}</legend>
    <p class="muted">{t('quick.prefixes_hint')}</p>
    <label class="field"
      ><span>{t('quick.prefix_todo')}</span><input class="input" bind:value={todo} maxlength="4" /></label
    >
    <label class="field"
      ><span>{t('quick.prefix_note')}</span><input class="input" bind:value={note} maxlength="4" /></label
    >
    <label class="field"
      ><span>{t('quick.prefix_bookmark')}</span><input
        class="input"
        bind:value={bookmark}
        maxlength="4"
      /></label
    >
  </fieldset>
  <button class="btn btn-primary" type="submit">{t('search.save')}</button>
</form>
