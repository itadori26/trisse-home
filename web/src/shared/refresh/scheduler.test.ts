import { describe, expect, it } from 'vitest';
import { defaultOptions, effectiveDelay, RefreshScheduler, type JobState } from './scheduler';

class FakeClock {
  t = 0;
  now() {
    return this.t;
  }
}

class FakeTicker {
  fn: (() => void) | null = null;
  every(_ms: number, fn: () => void) {
    this.fn = fn;
    return () => (this.fn = null);
  }
}

class FakeVisibility {
  shown = true;
  listeners: (() => void)[] = [];
  visible() {
    return this.shown;
  }
  onChange(fn: () => void) {
    this.listeners.push(fn);
    return () => {};
  }
  set(v: boolean) {
    this.shown = v;
    this.listeners.forEach((l) => l());
  }
}

const flush = () => new Promise((r) => setTimeout(r, 0));

function setup() {
  const clock = new FakeClock();
  const ticker = new FakeTicker();
  const vis = new FakeVisibility();
  const s = new RefreshScheduler(clock, ticker, vis, { random: () => 0.5 });
  const advance = async (ms: number) => {
    for (let i = 0; i < ms / 1000; i++) {
      clock.t += 1000;
      ticker.fn?.();
      await flush();
    }
  };
  return { clock, ticker, vis, s, advance };
}

describe('refresh scheduler (#117)', () => {
  it('runs at once and then every interval', async () => {
    const { s, advance } = setup();
    let runs = 0;
    s.register(async () => runs++, 10_000);
    await flush();
    expect(runs).toBe(1);
    await advance(9_000);
    expect(runs).toBe(1);
    await advance(2_000);
    expect(runs).toBe(2);
  });

  it('pauses while hidden and refreshes stale jobs when visible again', async () => {
    const { s, vis, advance } = setup();
    let runs = 0;
    s.register(async () => runs++, 10_000);
    await flush();
    vis.set(false);
    await advance(60_000);
    expect(runs).toBe(1);
    vis.set(true);
    await flush();
    expect(runs).toBe(2);
  });

  it('refreshes off-screen widgets less often', async () => {
    const { s, advance } = setup();
    let runs = 0;
    const h = s.register(async () => runs++, 10_000);
    await flush();
    h.setInViewport(false);
    await advance(30_000);
    expect(runs).toBe(1);
    await advance(11_000);
    expect(runs).toBe(2);
  });

  it('backs off after failures and resets on success', async () => {
    const { s, advance } = setup();
    let fail = true;
    let runs = 0;
    s.register(async () => {
      runs++;
      if (fail) throw new Error('down');
    }, 10_000);
    await flush(); // run 1 fails -> 20 s
    await advance(15_000);
    expect(runs).toBe(1);
    await advance(6_000); // run 2 fails -> 40 s
    expect(runs).toBe(2);
    fail = false;
    await advance(41_000); // run 3 succeeds -> back to 10 s
    expect(runs).toBe(3);
    await advance(11_000);
    expect(runs).toBe(4);
  });

  it('stops the shared timer when the last job is gone', async () => {
    const { s, ticker } = setup();
    const a = s.register(async () => {}, 1000);
    const b = s.register(async () => {}, 1000);
    a.dispose();
    expect(ticker.fn).not.toBeNull();
    b.dispose();
    expect(ticker.fn).toBeNull();
    expect(s.size).toBe(0);
  });

  it('caps the backoff and adds jitter', () => {
    const job: JobState = { intervalMs: 60_000, lastRun: 0, failures: 10, inViewport: true, running: false };
    expect(effectiveDelay(job, { ...defaultOptions, random: () => 0 })).toBe(15 * 60_000 * 0.9);
    expect(effectiveDelay(job, { ...defaultOptions, random: () => 0.999999 })).toBeCloseTo(
      15 * 60_000 * 1.1,
      -2,
    );
  });
});
