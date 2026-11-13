import { describe, expect, it } from 'vitest';
import {
  applyConfig,
  defaultFocusConfig as cfg,
  formatClock,
  initialState,
  parseFocusConfig,
  parseState,
  pause,
  remaining,
  reset,
  skip,
  start,
  tick,
} from './focus';

const MIN = 60_000;

describe('timer', () => {
  it('runs on timestamps and pauses with the time left', () => {
    let s = start(initialState(cfg), 0);
    expect(remaining(s, 10 * MIN)).toBe(15 * MIN);
    s = pause(s, 10 * MIN);
    expect(remaining(s, 99 * MIN)).toBe(15 * MIN);
    s = start(s, 100 * MIN);
    expect(s.endsAt).toBe(115 * MIN);
    expect(reset(s, cfg)).toMatchObject({ endsAt: null, left: 25 * MIN });
  });

  it('finishes a phase once, even after the tab slept for hours', () => {
    const s = start(initialState(cfg), 0);
    expect(tick(s, cfg, 24 * MIN).finished).toBeNull();
    const { state, finished } = tick(s, cfg, 300 * MIN);
    expect(finished).toEqual({ phase: 'work', at: 25 * MIN });
    expect(state).toEqual({ phase: 'short', done: 1, endsAt: null, left: 5 * MIN, total: 5 * MIN });
  });

  it('takes a long break after every n-th session and starts a new cycle', () => {
    let s = initialState({ ...cfg, longEvery: 2 });
    const c = { ...cfg, longEvery: 2 };
    const phases: string[] = [];
    let now = 0;
    for (let i = 0; i < 5; i++) {
      s = start(s, now);
      now = s.endsAt ?? now;
      s = tick(s, c, now).state;
      phases.push(`${s.phase}:${s.done}`);
    }
    expect(phases).toEqual(['short:1', 'work:1', 'long:2', 'work:0', 'short:1']);
  });

  it('follows changed durations only while a phase is untouched', () => {
    const short = { ...cfg, work: 1 };
    expect(applyConfig(initialState(cfg), short)).toMatchObject({ left: MIN, total: MIN });
    const started = pause(start(initialState(cfg), 0), 10 * MIN); // 15 min left
    expect(applyConfig(started, { ...cfg, work: 50 })).toEqual(started);
    expect(applyConfig(started, { ...cfg, work: 5 })).toMatchObject({ left: 5 * MIN });
    const running = start(initialState(cfg), 0);
    expect(applyConfig(running, short)).toBe(running);
  });

  it('skips without counting a session', () => {
    const s = skip(initialState(cfg), cfg);
    expect(s).toMatchObject({ phase: 'short', done: 0 });
    expect(skip(s, cfg)).toMatchObject({ phase: 'work', done: 0 });
  });
});

describe('formatClock', () => {
  it('rounds up to whole seconds', () => {
    expect(formatClock(25 * MIN)).toBe('25:00');
    expect(formatClock(61_500)).toBe('1:02');
    expect(formatClock(1)).toBe('0:01');
    expect(formatClock(0)).toBe('0:00');
  });
});

describe('parsing', () => {
  it('clamps the configuration', () => {
    expect(parseFocusConfig({})).toEqual(cfg);
    expect(parseFocusConfig({ work: 999, short: 0, longEvery: 2.5, notify: true })).toMatchObject({
      work: 180,
      short: 1,
      longEvery: 4,
      notify: true,
    });
  });

  it('repairs stored state', () => {
    expect(parseState(null, cfg)).toEqual(initialState(cfg));
    expect(parseState({ phase: 'nap' }, cfg)).toEqual(initialState(cfg));
    expect(parseState({ phase: 'short', done: 2, endsAt: null, left: 99 * MIN }, cfg)).toEqual({
      phase: 'short',
      done: 2,
      endsAt: null,
      left: 5 * MIN,
      total: 5 * MIN,
    });
  });
});
