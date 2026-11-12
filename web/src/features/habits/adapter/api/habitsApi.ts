import type { HttpClient } from '../../../../shared/api-client/http';
import type { HabitDTO } from '../../../../shared/api-client/generated/habits';
import type { HabitInputDTO, ToggleDTO } from '../../../../shared/api-client/generated/habits-requests';
import type { Habit, HabitInput } from '../../domain/habits';
import type { HabitsGateway } from '../../usecase/habitsUseCases';

const toHabit = (d: HabitDTO): Habit => ({
  id: d.id,
  name: d.name,
  emoji: d.emoji,
  days: d.days ?? [],
  doneToday: d.doneToday,
  current: d.current,
  best: d.best,
  heatmap: (d.heatmap ?? []).map((c) => ({ date: c.date, due: c.due, done: c.done })),
});

const toBody = (i: HabitInput): HabitInputDTO => ({ name: i.name, emoji: i.emoji, days: i.days });
const path = (id: string) => `/api/habits/${encodeURIComponent(id)}`;

/** HabitsGateway backed by the REST API. */
export class HabitsApi implements HabitsGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<Habit[]> {
    return (await this.http.get<HabitDTO[]>('/api/habits')).map(toHabit);
  }

  async create(input: HabitInput): Promise<Habit> {
    return toHabit(await this.http.post<HabitDTO>('/api/habits', toBody(input)));
  }

  async update(id: string, input: HabitInput): Promise<Habit> {
    return toHabit(await this.http.put<HabitDTO>(path(id), toBody(input)));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }

  async toggle(id: string, date: string): Promise<Habit> {
    const body: ToggleDTO = { date };
    return toHabit(await this.http.post<HabitDTO>(`${path(id)}/toggle`, body));
  }
}
