import { editableDays, type Habit, type HabitInput } from '../domain/habits';

/** Output port: habits backend. */
export interface HabitsGateway {
  list(): Promise<Habit[]>;
  create(input: HabitInput): Promise<Habit>;
  update(id: string, input: HabitInput): Promise<Habit>;
  remove(id: string): Promise<void>;
  toggle(id: string, date: string): Promise<Habit>;
}

const clean = (i: HabitInput): HabitInput => ({
  name: i.name.trim(),
  emoji: i.emoji.trim(),
  days: [...new Set(i.days)].sort(),
});

/** Application logic of the habit tracker (#106). */
export class HabitsUseCases {
  constructor(
    private readonly gateway: HabitsGateway,
    private readonly now: () => Date = () => new Date(),
  ) {}

  list(): Promise<Habit[]> {
    return this.gateway.list();
  }

  /** Creates (no id) or updates a habit. */
  save(id: string | null, input: HabitInput): Promise<Habit> {
    return id ? this.gateway.update(id, clean(input)) : this.gateway.create(clean(input));
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  /** Checks today off or undoes it. */
  toggleToday(id: string): Promise<Habit> {
    return this.gateway.toggle(id, editableDays(this.now()).today);
  }

  /** Checks yesterday off or undoes it (forgot to tick last night). */
  toggleYesterday(id: string): Promise<Habit> {
    return this.gateway.toggle(id, editableDays(this.now()).yesterday);
  }
}
