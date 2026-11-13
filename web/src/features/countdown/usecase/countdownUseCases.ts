import { type Countdown, type CountdownInput } from '../domain/countdown';

/** Output port: countdown backend. */
export interface CountdownGateway {
  list(graceDays: number): Promise<Countdown[]>;
  create(input: CountdownInput): Promise<Countdown>;
  update(id: string, input: CountdownInput): Promise<Countdown>;
  remove(id: string): Promise<void>;
}

const clean = (i: CountdownInput): CountdownInput => ({
  title: i.title.trim(),
  emoji: i.emoji.trim(),
  date: i.date,
  time: i.time,
});

/** Application logic of the countdown feature (#54). */
export class CountdownUseCases {
  constructor(private readonly gateway: CountdownGateway) {}

  /** Countdowns sorted by their end; past ones stay for graceDays. */
  list(graceDays: number): Promise<Countdown[]> {
    return this.gateway.list(graceDays);
  }

  /** Creates (no id) or updates a countdown. */
  save(id: string | null, input: CountdownInput): Promise<Countdown> {
    return id ? this.gateway.update(id, clean(input)) : this.gateway.create(clean(input));
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
}
