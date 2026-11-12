/**
 * Shared refresh scheduler for all widgets (#117): one timer instead of one
 * per widget; nothing runs while the page is hidden, stale jobs run at once
 * when it becomes visible again, widgets outside the viewport refresh less
 * often and failures back off exponentially with jitter.
 */

/** Current time in milliseconds (port; injected for tests). */
export interface Clock {
  now(): number;
}

/** Periodic callback (port); returns a function that stops it. */
export interface Ticker {
  every(ms: number, fn: () => void): () => void;
}

/** Page visibility (port). */
export interface VisibilitySource {
  visible(): boolean;
  /** Calls fn on every change; returns a function that unsubscribes. */
  onChange(fn: () => void): () => void;
}

/** Tuning of the scheduler. */
export interface SchedulerOptions {
  /** Resolution of the shared timer. */
  tickMs: number;
  /** Interval multiplier for widgets outside the viewport. */
  offscreenFactor: number;
  /** Upper bound of the backoff after failures. */
  maxBackoffMs: number;
  /** Random source in [0, 1) for jitter. */
  random: () => number;
}

export const defaultOptions: SchedulerOptions = {
  tickMs: 1000,
  offscreenFactor: 4,
  maxBackoffMs: 15 * 60_000,
  random: Math.random,
};

/** State of one registered job. */
export interface JobState {
  intervalMs: number;
  lastRun: number; // -Infinity before the first run
  failures: number;
  inViewport: boolean;
  running: boolean;
}

/**
 * Effective delay until the next run: the interval, stretched off-screen
 * and, after failures, doubled per failure up to maxBackoffMs, with ±10 %
 * jitter so many failing widgets do not retry in lockstep.
 */
export function effectiveDelay(job: JobState, o: SchedulerOptions): number {
  let ms = job.intervalMs * (job.inViewport ? 1 : o.offscreenFactor);
  if (job.failures > 0) {
    ms = Math.min(job.intervalMs * 2 ** job.failures, o.maxBackoffMs);
    ms *= 0.9 + o.random() * 0.2;
  }
  return ms;
}

/** Whether a job should run now. */
export function isDue(job: JobState, now: number, o: SchedulerOptions): boolean {
  return !job.running && now - job.lastRun >= effectiveDelay(job, o);
}

/** Handle of a registered job. */
export interface RefreshHandle {
  /** Stops the job. */
  dispose(): void;
  /** Reports whether the widget is visible in the viewport. */
  setInViewport(visible: boolean): void;
  /** Runs the job now (e.g. after the user changed a filter). */
  refreshNow(): void;
  /** Data freshness of the job (#94). */
  status(): JobStatus;
  /** Calls fn after every run; returns a function that unsubscribes. */
  onSettled(fn: () => void): () => void;
}

/** Freshness facts of a job (#94). */
export interface JobStatus {
  fetchedAt: number | null;
  lastError: string | null;
  intervalMs: number;
}

interface Job extends JobState {
  run: () => Promise<unknown>;
  fetchedAt: number | null;
  lastError: string | null;
  listeners: Set<() => void>;
}

const message = (err: unknown) => (err instanceof Error ? err.message : String(err));

/** The shared scheduler. */
export class RefreshScheduler {
  private readonly jobs = new Set<Job>();
  private stopTicker: (() => void) | null = null;
  private readonly o: SchedulerOptions;

  constructor(
    private readonly clock: Clock,
    private readonly ticker: Ticker,
    private readonly visibility: VisibilitySource,
    options: Partial<SchedulerOptions> = {},
  ) {
    this.o = { ...defaultOptions, ...options };
    // Back on screen: refresh everything that went stale while hidden.
    visibility.onChange(() => {
      if (this.visibility.visible()) this.tick();
    });
  }

  /**
   * Registers a job; it runs at once (if the page is visible) and then
   * every intervalMs. A rejected promise counts as a failure.
   */
  register(run: () => Promise<unknown>, intervalMs: number): RefreshHandle {
    const job: Job = {
      run,
      intervalMs,
      lastRun: -Infinity,
      failures: 0,
      inViewport: true,
      running: false,
      fetchedAt: null,
      lastError: null,
      listeners: new Set(),
    };
    this.jobs.add(job);
    this.stopTicker ??= this.ticker.every(this.o.tickMs, () => this.tick());
    if (this.visibility.visible()) this.start(job);
    return {
      dispose: () => {
        this.jobs.delete(job);
        if (this.jobs.size === 0 && this.stopTicker) {
          this.stopTicker();
          this.stopTicker = null;
        }
      },
      setInViewport: (v) => {
        const wasOff = !job.inViewport;
        job.inViewport = v;
        // Scrolled into view after being stale: refresh now.
        if (v && wasOff && this.visibility.visible() && isDue(job, this.clock.now(), this.o)) this.start(job);
      },
      refreshNow: () => {
        if (!job.running) this.start(job);
      },
      status: () => ({ fetchedAt: job.fetchedAt, lastError: job.lastError, intervalMs: job.intervalMs }),
      onSettled: (fn) => {
        job.listeners.add(fn);
        return () => job.listeners.delete(fn);
      },
    };
  }

  /** Runs all due jobs; nothing happens while the page is hidden. */
  tick(): void {
    if (!this.visibility.visible()) return;
    const now = this.clock.now();
    for (const job of this.jobs) {
      if (isDue(job, now, this.o)) this.start(job);
    }
  }

  /** Number of registered jobs (for tests and diagnostics). */
  get size(): number {
    return this.jobs.size;
  }

  private start(job: Job): void {
    job.running = true;
    job.lastRun = this.clock.now();
    job
      .run()
      .then(
        () => {
          job.failures = 0;
          job.fetchedAt = this.clock.now();
          job.lastError = null;
        },
        (err: unknown) => {
          job.failures = Math.min(job.failures + 1, 20);
          job.lastError = message(err);
        },
      )
      .finally(() => {
        job.running = false;
        for (const fn of job.listeners) fn();
        // Measure the interval from the end of a slow request.
        job.lastRun = this.clock.now();
      });
  }
}
