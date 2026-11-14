<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    emptyModule,
    formatEcts,
    formatGrade,
    type ModuleInput,
    STATUSES,
    type StudyModule,
  } from '../domain/study';
  import type { StudyUseCases } from '../usecase/studyUseCases';

  /** Settings → Study progress: total ECTS and the modules (#57). */
  interface Props {
    study: StudyUseCases;
  }
  let { study }: Props = $props();

  let total = $state(180);
  let modules = $state<StudyModule[]>([]);
  /** The module being edited: its id, 'new' or null. */
  let editing = $state<string | null>(null);
  let form = $state<ModuleInput>(emptyModule());
  let error = $state('');
  let message = $state('');

  const load = () =>
    Promise.all([study.totalEcts(), study.modules()])
      .then(([v, list]) => ((total = v), (modules = list)))
      .catch((err) => (error = errorMessage(err)));
  $effect(() => {
    void load();
  });

  async function run(action: () => Promise<unknown>) {
    error = message = '';
    try {
      await action();
      message = t('study.saved');
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function edit(m?: StudyModule) {
    editing = m ? m.id : 'new';
    form = m
      ? { name: m.name, ects: m.ects, semester: m.semester, status: m.status, grade: m.grade }
      : emptyModule();
    message = '';
  }

  function saveModule(e: SubmitEvent) {
    e.preventDefault();
    const id = editing === 'new' ? undefined : (editing ?? undefined);
    void run(async () => {
      await study.save(form, id);
      editing = null;
    });
  }

  function remove(m: StudyModule) {
    if (!confirm(t('study.delete_confirm', { name: m.name }))) return;
    void run(() => study.remove(m.id));
  }

  const statusLabel = (s: string) => t(`study.status_${s}` as 'study.status_open');
  // A grade input bound to number | null: empty means no grade.
  const gradeValue = (v: string) => (v.trim() === '' ? null : Number(v.replace(',', '.')));
</script>

<p class="muted">{t('study.intro')}</p>
<form
  class="total"
  onsubmit={(e) => {
    e.preventDefault();
    void run(() => study.setTotalEcts(Number(total)));
  }}
>
  <label class="field"
    ><span>{t('study.total_ects')}</span><input
      class="input"
      type="number"
      min="1"
      max="600"
      step="0.5"
      bind:value={total}
    /></label
  >
  <button class="btn" type="submit">{t('study.save')}</button>
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

{#snippet editor()}
  <form class="editor" onsubmit={saveModule}>
    <label class="field name"
      ><span>{t('study.name')}</span><input
        class="input"
        bind:value={form.name}
        maxlength="80"
        required
      /></label
    >
    <label class="field"
      ><span>{t('study.ects')}</span><input
        class="input"
        type="number"
        min="0.5"
        max="60"
        step="0.5"
        bind:value={form.ects}
      /></label
    >
    <label class="field"
      ><span>{t('study.semester')}</span><select class="input" bind:value={form.semester}>
        <option value={0}>{t('study.semester_none')}</option>
        {#each Array.from({ length: 20 }, (_, i) => i + 1) as s (s)}<option value={s}>{s}</option>{/each}
      </select></label
    >
    <label class="field"
      ><span>{t('study.status')}</span><select class="input" bind:value={form.status}>
        {#each STATUSES as s (s)}<option value={s}>{statusLabel(s)}</option>{/each}
      </select></label
    >
    {#if form.status === 'passed'}
      <label class="field"
        ><span>{t('study.grade')}</span><input
          class="input"
          inputmode="decimal"
          value={form.grade ?? ''}
          placeholder={t('study.grade_none')}
          oninput={(e) => (form.grade = gradeValue(e.currentTarget.value))}
        /></label
      >
    {/if}
    <div class="actions">
      <button class="btn btn-primary" type="submit">{t('study.save')}</button>
      <button class="btn" type="button" onclick={() => (editing = null)}>{t('study.cancel')}</button>
    </div>
  </form>
{/snippet}

<h3>{t('study.modules')}</h3>
<table>
  <thead>
    <tr>
      <th>{t('study.name')}</th>
      <th>{t('study.ects')}</th>
      <th>{t('study.semester')}</th>
      <th>{t('study.status')}</th>
      <th>{t('study.grade')}</th>
      <th></th>
    </tr>
  </thead>
  <tbody>
    {#each modules as m (m.id)}
      {#if editing === m.id}
        <tr><td colspan="6">{@render editor()}</td></tr>
      {:else}
        <tr class={m.status}>
          <td>{m.name}</td>
          <td>{formatEcts(m.ects, intlLocale())}</td>
          <td>{m.semester || t('study.semester_none')}</td>
          <td>{statusLabel(m.status)}</td>
          <td>{m.grade !== null ? formatGrade(m.grade, intlLocale()) : ''}</td>
          <td class="buttons">
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('study.edit', { name: m.name })}
              onclick={() => edit(m)}>✎</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('study.delete', { name: m.name })}
              onclick={() => remove(m)}>✕</button
            >
          </td>
        </tr>
      {/if}
    {:else}
      <tr><td colspan="6" class="muted">{t('study.no_modules')}</td></tr>
    {/each}
  </tbody>
</table>
{#if editing === 'new'}
  {@render editor()}
{:else}
  <button class="btn btn-primary" type="button" onclick={() => edit()}>{t('study.add')}</button>
{/if}

<style>
  .total {
    display: flex;
    align-items: flex-end;
    gap: var(--space-2);
  }
  h3 {
    margin: var(--space-3) 0 var(--space-1);
    font-size: 1em;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: var(--space-2);
  }
  th,
  td {
    text-align: left;
    padding: 2px var(--space-1);
    border-bottom: 1px solid var(--color-border);
  }
  th {
    font-size: 0.8em;
    color: var(--color-text-muted);
  }
  tr.passed td:first-child {
    font-weight: 600;
  }
  tr.open td {
    color: var(--color-text-muted);
  }
  .buttons {
    white-space: nowrap;
    text-align: right;
  }
  .editor {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: flex-end;
    padding: var(--space-2) 0;
  }
  .editor .field {
    flex: 1 1 6rem;
    min-width: 0;
  }
  .editor .name {
    flex: 3 1 14rem;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
  @media (max-width: 600px) {
    th:nth-child(3),
    td:nth-child(3) {
      display: none;
    }
  }
</style>
