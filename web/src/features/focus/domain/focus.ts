/**
 * Focus timer (#58): a pure state machine work → short break → … → long
 * break. It is based on timestamps, not on counting ticks, so it keeps the
 * right time while the browser tab sleeps.
 */

export type Phase = 'work' | 'short' | 'long';

/** Per-widget settings; durations in minutes. */
export interface FocusConfig {
  work: number;
  short: number;
  long: number;
  /** A long break follows every n-th focus session. */
  longEvery: number;
  /** Show a browser notification at the end of a phase. */
  notify: boolean;
}

export const defaultFocusConfig: FocusConfig = { work: 25, short: 5, long: 15, longEvery: 4, notify: false };

/** Upper limit of a phase; mirrors the backend. */
export const MAX_MINUTES = 180;

const int = (v: unknown, d: number, lo: number, hi: number) =>
  typeof v === 'number' && Number.isInteger(v) ? Math.min(Math.max(v, lo), hi) : d;

/** Reads the widget configuration, falling back to defaults. */
export function parseFocusConfig(raw: Record<string, unknown>): FocusConfig {
  const d = defaultFocusConfig;
  return {
    work: int(raw['work'], d.work, 1, MAX_MINUTES),
    short: int(raw['short'], d.short, 1, MAX_MINUTES),
    long: int(raw['long'], d.long, 1, MAX_MINUTES),
    longEvery: int(raw['longEvery'], d.longEvery, 1, 12),
    notify: typeof raw['notify'] === 'boolean' ? raw['notify'] : d.notify,
  };
}

/** State of the timer; plain data so it can be stored. */
export interface TimerState {
  phase: Phase;
  /** Focus sessions finished in the current cycle (until the long break). */
  done: number;
  /** End of the phase in epoch ms while running, otherwise null. */
  endsAt: number | null;
  /** Time left in ms while paused. */
  left: number;
  /** Full length of the current phase in ms when it was prepared. */
  total: number;
}

const MINUTE = 60_000;

export const durationOf = (phase: Phase, cfg: FocusConfig): number => cfg[phase] * MINUTE;

/** A paused, untouched phase. */
function fresh(phase: Phase, done: number, cfg: FocusConfig): TimerState {
  const d = durationOf(phase, cfg);
  return { phase, done, endsAt: null, left: d, total: d };
}

export function initialState(cfg: FocusConfig): TimerState {
  return fresh('work', 0, cfg);
}

/**
 * Follows changed settings: a paused phase that was not started yet takes
 * the new length; a started one keeps its time, capped at the new length.
 */
export function applyConfig(s: TimerState, cfg: FocusConfig): TimerState {
  if (isRunning(s)) return s;
  const d = durationOf(s.phase, cfg);
  if (s.left === s.total) return s.total === d ? s : fresh(s.phase, s.done, cfg);
  return s.left > d ? { ...s, left: d, total: d } : s;
}

export const isRunning = (s: TimerState): boolean => s.endsAt !== null;

/** Milliseconds left at now (never negative). */
export function remaining(s: TimerState, now: number): number {
  return s.endsAt === null ? s.left : Math.max(0, s.endsAt - now);
}

export function start(s: TimerState, now: number): TimerState {
  return isRunning(s) ? s : { ...s, endsAt: now + s.left };
}

export function pause(s: TimerState, now: number): TimerState {
  return isRunning(s) ? { ...s, endsAt: null, left: remaining(s, now) } : s;
}

/** Starts the current phase over (paused). */
export function reset(s: TimerState, cfg: FocusConfig): TimerState {
  return fresh(s.phase, s.done, cfg);
}

/** The phase after the current one; done counts the finished sessions. */
function next(phase: Phase, done: number, cfg: FocusConfig): { phase: Phase; done: number } {
  if (phase !== 'work') return { phase: 'work', done: phase === 'long' ? 0 : done };
  const d = done + 1;
  return { phase: d % cfg.longEvery === 0 ? 'long' : 'short', done: d };
}

/** Jumps to the next phase without counting a session. */
export function skip(s: TimerState, cfg: FocusConfig): TimerState {
  const n = s.phase === 'work' ? { phase: 'short' as Phase, done: s.done } : next(s.phase, s.done, cfg);
  return fresh(n.phase, n.done, cfg);
}

/** What happened when a running phase ran out. */
export interface Finished {
  phase: Phase;
  /** When it actually ended (may be in the past after the tab slept). */
  at: number;
}

/**
 * Advances a running timer whose phase has ended: the next phase is
 * prepared, paused, so a sleeping tab never skips several phases.
 */
export function tick(
  s: TimerState,
  cfg: FocusConfig,
  now: number,
): { state: TimerState; finished: Finished | null } {
  if (s.endsAt === null || now < s.endsAt) return { state: s, finished: null };
  const n = next(s.phase, s.done, cfg);
  return {
    state: fresh(n.phase, n.done, cfg),
    finished: { phase: s.phase, at: s.endsAt },
  };
}

/** mm:ss of a duration in ms (rounded up, so 0:01 is shown until the end). */
export function formatClock(ms: number): string {
  const total = Math.ceil(ms / 1000);
  const m = Math.floor(total / 60);
  return `${m}:${String(total % 60).padStart(2, '0')}`;
}

/** Validates a stored state (localStorage may hold anything). */
export function parseState(raw: unknown, cfg: FocusConfig): TimerState {
  if (typeof raw !== 'object' || raw === null) return initialState(cfg);
  const r = raw as Record<string, unknown>;
  const phase = r['phase'];
  if (phase !== 'work' && phase !== 'short' && phase !== 'long') return initialState(cfg);
  const num = (v: unknown) => (typeof v === 'number' && Number.isFinite(v) ? v : null);
  const total = Math.min(Math.max(num(r['total']) ?? durationOf(phase, cfg), 0), MAX_MINUTES * MINUTE);
  const left = num(r['left']);
  return {
    phase,
    done: Math.max(0, Math.floor(num(r['done']) ?? 0)),
    endsAt: num(r['endsAt']),
    left: left === null ? total : Math.min(Math.max(left, 0), total),
    total,
  };
}

/** One day of the weekly statistics. */
export interface FocusDay {
  /** YYYY-MM-DD */
  date: string;
  sessions: number;
  minutes: number;
}
