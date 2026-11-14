/**
 * Polite screen reader announcements (#113). All widgets share one live
 * region; the announcer drops a message that was already announced in the
 * last minutes, so polling widgets that keep failing cannot flood it.
 */
export class Announcer {
  private readonly recent = new Map<string, number>();

  constructor(
    private readonly sink: (message: string) => void,
    private readonly now: () => number = Date.now,
    private readonly repeatAfterMs = 5 * 60_000,
  ) {}

  /** Announces message unless it was announced recently; returns whether it was. */
  announce(message: string): boolean {
    const at = this.now();
    const last = this.recent.get(message);
    if (message === '' || (last !== undefined && at - last < this.repeatAfterMs)) return false;
    for (const [m, t] of this.recent) if (at - t >= this.repeatAfterMs) this.recent.delete(m);
    this.recent.set(message, at);
    this.sink(message);
    return true;
  }
}
