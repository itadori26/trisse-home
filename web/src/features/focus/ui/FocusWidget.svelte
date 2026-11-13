<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { untrack } from 'svelte';
  import {
    applyConfig,
    type FocusDay,
    formatClock,
    initialState,
    isRunning,
    parseFocusConfig,
    parseState,
    pause,
    type Phase,
    remaining,
    reset,
    skip,
    start,
    tick,
    type TimerState,
  } from '../domain/focus';
  import type { FocusUseCases } from '../usecase/focusUseCases';

  /** Props of the focus timer widget (#58). */
  interface Props {
    focus: FocusUseCases;
    config: Record<string, unknown>;
  }
  let { focus, config }: Props = $props();

  const STORAGE_KEY = 'trisse-focus-timer';
  const cfg = $derived(parseFocusConfig(config));

  function restore(): TimerState {
    const c = parseFocusConfig(config);
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      return raw ? parseState(JSON.parse(raw), c) : initialState(c);
    } catch {
      return initialState(c);
    }
  }

  let timer = $state<TimerState>(restore());
  let now = $state(Date.now());
  let week = $state<FocusDay[]>([]);
  let error = $state('');

  function save(s: TimerState) {
    timer = s;
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
    } catch {
      // Private mode: the timer still works until the page is closed.
    }
  }

  // Changed durations apply to a phase that has not been started yet.
  $effect(() => {
    const next = applyConfig(
      untrack(() => timer),
      cfg,
    );
    if (next !== untrack(() => timer)) save(next);
  });

  const phaseLabel = (p: Phase) => t(`focus.phase.${p}`);

  async function finished(phase: Phase, at: number) {
    focus.notify(cfg.notify, t('focus.notify_title', { phase: phaseLabel(phase) }), t('focus.notify_body'));
    if (phase !== 'work') return;
    try {
      week = await focus.record(cfg.work, new Date(at));
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function step() {
    now = Date.now();
    const r = tick(timer, cfg, now);
    if (r.finished) {
      save(r.state);
      void finished(r.finished.phase, r.finished.at);
    }
  }

  // A one-second tick only while running; the end is computed from the
  // stored timestamp, so a throttled background tab still ends correctly.
  $effect(() => {
    if (!isRunning(timer)) return;
    step();
    const id = setInterval(step, 1000);
    document.addEventListener('visibilitychange', step);
    return () => {
      clearInterval(id);
      document.removeEventListener('visibilitychange', step);
    };
  });

  $effect(() => {
    focus
      .stats()
      .then((d) => (week = d))
      .catch((err) => (error = errorMessage(err)));
  });

  const left = $derived(remaining(timer, now));
  const max = $derived(Math.max(1, ...week.map((d) => d.sessions)));
  const weekday = (date: string) =>
    new Date(`${date}T12:00:00`).toLocaleDateString(intlLocale(), { weekday: 'narrow' });
  const total = $derived(week.reduce((n, d) => n + d.sessions, 0));
</script>

<div class="focus" class:running={isRunning(timer)} data-phase={timer.phase}>
  <p class="phase">{phaseLabel(timer.phase)}</p>
  <p class="clock" role="timer" aria-live="off">{formatClock(left)}</p>
  <p class="dots" aria-label={t('focus.cycle', { done: timer.done % cfg.longEvery, of: cfg.longEvery })}>
    {#each Array.from({ length: cfg.longEvery }, (_, i) => i) as i (i)}
      <span class:filled={i < timer.done % cfg.longEvery || (timer.phase === 'long' && timer.done > 0)}
      ></span>
    {/each}
  </p>
  <div class="controls">
    {#if isRunning(timer)}
      <button class="btn btn-primary" type="button" onclick={() => save(pause(timer, Date.now()))}
        >{t('focus.pause')}</button
      >
    {:else}
      <button class="btn btn-primary" type="button" onclick={() => save(start(timer, Date.now()))}
        >{t('focus.start')}</button
      >
    {/if}
    <button class="btn" type="button" onclick={() => save(reset(timer, cfg))}>{t('focus.reset')}</button>
    <button class="btn" type="button" onclick={() => save(skip(timer, cfg))}>{t('focus.skip')}</button>
  </div>
  {#if error}<p class="error">{error}</p>{/if}
  {#if week.length > 0}
    <figure class="week" aria-label={t('focus.week_total', { n: total })}>
      {#each week as d (d.date)}
        <div class="day" title={t('focus.day_title', { n: d.sessions, minutes: d.minutes })}>
          <span class="bar" style:height="{(d.sessions / max) * 100}%"></span>
          <span class="label">{weekday(d.date)}</span>
        </div>
      {/each}
      <figcaption class="muted">{t('focus.week_total', { n: total })}</figcaption>
    </figure>
  {/if}
</div>

<style>
  .focus {
    display: grid;
    justify-items: center;
    gap: var(--space-1);
    text-align: center;
  }
  p {
    margin: 0;
  }
  .phase {
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--color-text-muted);
  }
  .clock {
    font-size: 2.6rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 1.1;
  }
  .running .clock {
    color: var(--color-accent);
  }
  [data-phase='short'] .clock,
  [data-phase='long'] .clock {
    color: var(--color-success, var(--color-accent));
  }
  .dots {
    display: flex;
    gap: 4px;
  }
  .dots span {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    border: 1px solid var(--color-text-muted);
  }
  .dots span.filled {
    background: var(--color-accent);
    border-color: var(--color-accent);
  }
  .controls {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-1);
  }
  .week {
    margin: var(--space-2) 0 0;
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 4px;
    width: 100%;
    max-width: 16rem;
  }
  .day {
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    align-items: center;
    height: 3rem;
  }
  .bar {
    width: 60%;
    min-height: 2px;
    background: var(--color-accent);
    border-radius: 2px 2px 0 0;
  }
  .label {
    font-size: 0.7rem;
    color: var(--color-text-muted);
  }
  figcaption {
    grid-column: 1 / -1;
    font-size: 0.75rem;
  }
</style>
