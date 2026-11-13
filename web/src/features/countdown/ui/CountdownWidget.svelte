<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import {
    type Countdown,
    isSoon,
    parseCountdownConfig,
    type Remaining,
    remainingAt,
    target,
  } from '../domain/countdown';
  import type { CountdownUseCases } from '../usecase/countdownUseCases';
  import CountdownDialog from './CountdownDialog.svelte';

  /** Props of the countdown widget (#54). */
  interface Props {
    countdown: CountdownUseCases;
    config: Record<string, unknown>;
  }
  let { countdown, config }: Props = $props();
  const poll = usePolling();

  const cfg = $derived(parseCountdownConfig(config));
  let items = $state<Countdown[]>([]);
  let error = $state('');
  let editing = $state<Countdown | null | undefined>(undefined); // null = new
  let now = $state(new Date());

  async function load() {
    try {
      items = await countdown.list(cfg.graceDays);
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => poll(load, 15 * 60_000));

  // The remaining time is computed locally, so a minute tick is enough.
  $effect(() => {
    const timer = setInterval(() => (now = new Date()), 60_000);
    return () => clearInterval(timer);
  });

  function label(r: Remaining): string {
    switch (r.kind) {
      case 'days':
        return t(r.days === 1 ? 'countdown.one_day' : 'countdown.days', { n: r.days });
      case 'hours':
        return t('countdown.hours', { h: r.hours, m: String(r.minutes).padStart(2, '0') });
      case 'today':
        return t('countdown.today');
      case 'past':
        return t('countdown.past');
    }
  }

  function when(c: Countdown): string {
    const opts: Intl.DateTimeFormatOptions = { day: 'numeric', month: 'short', year: 'numeric' };
    if (c.time) Object.assign(opts, { hour: '2-digit', minute: '2-digit' });
    return target(c).toLocaleString(intlLocale(), opts);
  }

  async function remove(c: Countdown) {
    if (!confirm(t('common.confirm_delete', { name: c.title }))) return;
    try {
      await countdown.remove(c.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error">{error}</p>{/if}
<ul>
  {#each items as c (c.id)}
    {@const r = remainingAt(c, now)}
    <li class:soon={isSoon(r, cfg.highlightDays)} class:past={r.kind === 'past'}>
      <span class="emoji" aria-hidden="true">{c.emoji || '⏳'}</span>
      <span class="main">
        <span class="title">{c.title}</span>
        <span class="when muted">{when(c)}</span>
      </span>
      <span class="left">{label(r)}</span>
      <span class="actions">
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('countdown.edit_named', { title: c.title })}
          onclick={() => (editing = c)}>✎</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('countdown.remove', { title: c.title })}
          onclick={() => remove(c)}>✕</button
        >
      </span>
    </li>
  {:else}
    <li class="muted">{t('countdown.empty')}</li>
  {/each}
</ul>
<button class="btn add" type="button" onclick={() => (editing = null)}>+ {t('countdown.new')}</button>

{#if editing !== undefined}
  <CountdownDialog
    {countdown}
    item={editing}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = undefined)}
  />
{/if}

<style>
  ul {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 0;
  }
  .emoji {
    font-size: 1.3rem;
    width: 1.6rem;
    text-align: center;
  }
  .main {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  .title,
  .when {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .when {
    font-size: 0.8rem;
  }
  .left {
    font-variant-numeric: tabular-nums;
    font-weight: 600;
    white-space: nowrap;
  }
  li.soon .left {
    color: var(--color-accent);
  }
  li.past {
    opacity: 0.6;
  }
  .actions {
    display: flex;
    opacity: 0;
    transition: opacity 0.15s;
  }
  li:hover .actions,
  li:focus-within .actions {
    opacity: 1;
  }
  @media (hover: none) {
    .actions {
      opacity: 1;
    }
  }
</style>
