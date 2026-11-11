<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import type { Tab, TabTemplate } from '../domain/dashboard';
  import type { DashboardUseCases } from '../usecase/dashboardUseCases';

  /** Props of the dialog that creates a tab, optionally from a template. */
  interface Props {
    uc: DashboardUseCases;
    oncreated: (tab: Tab) => void;
    onclose: () => void;
  }
  let { uc, oncreated, onclose }: Props = $props();

  let name = $state('');
  let templateId = $state('');
  let templates = $state<TabTemplate[]>([]);
  let error = $state('');

  $effect(() => {
    uc.listTemplates()
      .then((t) => (templates = t))
      .catch(() => {}); // the dialog works without templates
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      const tab = templateId ? await uc.createTabFromTemplate(templateId, name) : await uc.createTab(name);
      oncreated(tab);
      onclose();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(t: TabTemplate) {
    if (!confirm(`Delete the template "${t.name}"?`)) return;
    try {
      await uc.deleteTemplate(t.id);
      templates = templates.filter((x) => x.id !== t.id);
      if (templateId === t.id) templateId = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<Modal title="New tab" {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>Name {templateId ? '(empty = template name)' : ''}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={name} autofocus required={!templateId} maxlength="40" />
    </label>
    {#if templates.length}
      <fieldset>
        <legend>Start from</legend>
        <label><input type="radio" bind:group={templateId} value="" /> Empty tab</label>
        {#each templates as t (t.id)}
          <div class="tpl">
            <label
              ><input type="radio" bind:group={templateId} value={t.id} />
              {t.name} <span class="muted">({t.widgets} widgets)</span></label
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label="Delete template {t.name}"
              onclick={() => remove(t)}>✕</button
            >
          </div>
        {/each}
      </fieldset>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit">Create</button>
  </form>
</Modal>

<style>
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    display: grid;
    gap: var(--space-1);
    margin: var(--space-2) 0 var(--space-3);
  }
  .tpl {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
</style>
