<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type BriefingSettings, moveEntry, sectionTitle } from '../domain/briefing';
  import type { BriefingUseCases } from '../usecase/briefingUseCases';

  /** Props of the briefing settings: sections, daily push, location (#52, #70). */
  interface Props {
    briefing: BriefingUseCases;
    /** Whether the notifications module is enabled (push needs it). */
    canPush: boolean;
  }
  let { briefing, canPush }: Props = $props();

  const weekdays = [1, 2, 3, 4, 5, 6, 0];
  // 8 November 2026 is a Sunday, so 8 + d is weekday d.
  const dayName = (d: number) =>
    new Date(2026, 10, 8 + d).toLocaleDateString(intlLocale(), { weekday: 'short' });

  let form = $state<BriefingSettings | null>(null);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    briefing
      .settings()
      .then((s) => (form = s))
      .catch((err) => (error = errorMessage(err)));
  });

  async function run(action: () => Promise<void>, done: string) {
    error = message = '';
    busy = true;
    try {
      await action();
      message = done;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function save(e: SubmitEvent) {
    e.preventDefault();
    const current = form;
    if (!current) return;
    run(async () => {
      form = await briefing.saveSettings(current);
    }, t('common.saved'));
  }

  function move(i: number, delta: number) {
    if (!form) return;
    const next = moveEntry(form.layout, i, delta);
    if (next) form.layout = next;
  }
</script>

{#if form}
  <form onsubmit={save}>
    <fieldset>
      <legend>{t('briefing.sections')}</legend>
      <p class="muted small">{t('briefing.sections_hint')}</p>
      <ol class="layout">
        {#each form.layout as entry, i (entry.kind)}
          {@const available = form.available.includes(entry.kind)}
          <li class:off={!available}>
            <label>
              <input type="checkbox" bind:checked={entry.enabled} disabled={!available} />
              {sectionTitle(entry.kind)}
              {#if !available}<span class="muted small">({t('briefing.module_disabled')})</span>{/if}
            </label>
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('briefing.move_up', { name: sectionTitle(entry.kind) })}
              disabled={i === 0}
              onclick={() => move(i, -1)}>↑</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('briefing.move_down', { name: sectionTitle(entry.kind) })}
              disabled={i === form.layout.length - 1}
              onclick={() => move(i, 1)}>↓</button
            >
          </li>
        {/each}
      </ol>
    </fieldset>

    <fieldset>
      <legend>{t('briefing.location')}</legend>
      <p class="muted small">{t('briefing.location_hint')}</p>
      <div class="row">
        <label class="inline"
          >{t('briefing.lat')}
          <input
            class="input num"
            type="number"
            step="0.0001"
            min="-90"
            max="90"
            bind:value={form.lat}
          /></label
        >
        <label class="inline"
          >{t('briefing.lon')}
          <input
            class="input num"
            type="number"
            step="0.0001"
            min="-180"
            max="180"
            bind:value={form.lon}
          /></label
        >
      </div>
    </fieldset>

    <fieldset disabled={!canPush}>
      <legend>{t('briefing.push')}</legend>
      {#if !canPush}<p class="muted small">{t('briefing.push_needs_notifications')}</p>{/if}
      <label class="inline"
        ><input type="checkbox" bind:checked={form.pushOn} />
        {t('briefing.push_enable')}
        <input class="input" type="time" bind:value={form.pushAt} required /></label
      >
      <div class="row" role="group" aria-label={t('briefing.push_days')}>
        {#each weekdays as d (d)}
          <label
            ><input
              type="checkbox"
              checked={form.pushDays.includes(d)}
              onchange={(e) =>
                form &&
                (form.pushDays = e.currentTarget.checked
                  ? [...form.pushDays, d]
                  : form.pushDays.filter((x) => x !== d))}
            />
            {dayName(d)}</label
          >
        {/each}
      </div>
      <p class="muted small">{t('briefing.push_days_hint')}</p>
      <label class="inline"
        ><input type="checkbox" bind:checked={form.skipEmpty} /> {t('briefing.skip_empty')}</label
      >
    </fieldset>

    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if message}<p class="muted" role="status">{message}</p>{/if}
    <p class="actions">
      <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
      {#if canPush}
        <button
          class="btn"
          type="button"
          disabled={busy}
          onclick={() => run(() => briefing.sendNow(), t('briefing.sent'))}>{t('briefing.send_now')}</button
        >
      {/if}
    </p>
  </form>
{:else if error}
  <p class="error">{error}</p>
{/if}

<style>
  fieldset {
    margin: 0 0 var(--space-3);
    display: grid;
    gap: var(--space-2);
  }
  .layout {
    margin: 0;
    padding-left: 1.4em;
    display: grid;
    gap: 2px;
  }
  .layout li {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  .layout label {
    flex: 1;
  }
  .off {
    opacity: 0.6;
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1) var(--space-3);
    align-items: center;
  }
  .inline {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
  }
  .num {
    width: 8em;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
  .small {
    font-size: 0.85rem;
    margin: 0;
  }
</style>
