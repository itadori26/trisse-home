<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    type CalendarSubscription,
    emptySubscription,
    refreshChoices,
    type SubscriptionInput,
  } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';

  /** Props of the ICS subscription management (#98). */
  interface Props {
    calendar: CalendarUseCases;
  }
  let { calendar }: Props = $props();

  let list = $state<CalendarSubscription[]>([]);
  let form = $state<SubscriptionInput>(emptySubscription());
  let editing = $state<CalendarSubscription | null>(null);
  let error = $state('');
  let busy = $state(false);

  const refreshLabel = (m: number) =>
    m < 60 ? t('calendar.refresh_minutes', { n: m }) : t('calendar.refresh_hours', { n: m / 60 });

  async function run(action: () => Promise<void>) {
    error = '';
    busy = true;
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  $effect(() => {
    run(async () => {
      list = await calendar.subscriptions();
    });
  });

  function edit(s: CalendarSubscription) {
    editing = s;
    form = { name: s.name, url: '', color: s.color, refreshMinutes: s.refreshMinutes };
  }

  function reset() {
    editing = null;
    form = emptySubscription();
  }

  function submit(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      await calendar.saveSubscription(form, editing?.id);
      list = await calendar.subscriptions();
      reset();
    });
  }

  function remove(s: CalendarSubscription) {
    if (!confirm(t('common.confirm_delete', { name: s.name }))) return;
    run(async () => {
      await calendar.deleteSubscription(s.id);
      list = list.filter((x) => x.id !== s.id);
      if (editing?.id === s.id) reset();
    });
  }
</script>

<h4>{t('calendar.subscriptions')}</h4>
<p class="muted">{t('calendar.subscriptions_hint')}</p>
<ul class="list">
  {#each list as s (s.id)}
    <li>
      <span class="swatch" style:background={s.color} aria-hidden="true"></span>
      <span class="name"
        >{s.name} <small class="muted">{s.urlHint} · {refreshLabel(s.refreshMinutes)}</small></span
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('calendar.edit_subscription', { name: s.name })}
        onclick={() => edit(s)}>✎</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('calendar.delete_subscription', { name: s.name })}
        onclick={() => remove(s)}>✕</button
      >
    </li>
  {/each}
</ul>
<form onsubmit={submit}>
  <div class="row">
    <label class="field"
      ><span>{t('calendar.subscription_name')}</span><input
        class="input"
        bind:value={form.name}
        maxlength="60"
        required
      /></label
    >
    <label class="field"
      ><span>{t('calendar.subscription_color')}</span><input
        class="input"
        type="color"
        bind:value={form.color}
      /></label
    >
  </div>
  <label class="field"
    ><span
      >{editing
        ? t('calendar.subscription_url_keep', { hint: editing.urlHint })
        : t('calendar.subscription_url')}</span
    ><input
      class="input"
      bind:value={form.url}
      required={!editing}
      autocomplete="off"
      placeholder="webcal://example.com/calendar.ics"
    /></label
  >
  <label class="field"
    ><span>{t('calendar.subscription_refresh')}</span><select class="input" bind:value={form.refreshMinutes}>
      {#each refreshChoices as m (m)}<option value={m}>{refreshLabel(m)}</option>{/each}
    </select></label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p>
    <button class="btn btn-primary" type="submit" disabled={busy}
      >{editing ? t('common.save') : t('calendar.add_subscription')}</button
    >
    {#if editing}<button class="btn" type="button" onclick={reset}>{t('common.cancel')}</button>{/if}
  </p>
</form>

<style>
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .list li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .name {
    flex: 1;
    display: grid;
  }
  .swatch {
    width: 0.8em;
    height: 0.8em;
    border-radius: 50%;
    flex: none;
  }
  .row {
    display: grid;
    grid-template-columns: 1fr 6rem;
    gap: var(--space-2);
  }
</style>
