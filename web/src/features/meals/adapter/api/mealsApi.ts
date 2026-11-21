import type { HttpClient } from '../../../../shared/api-client/http';
import type { EntryDTO } from '../../../../shared/api-client/generated/meals';
import type { EntryInputDTO } from '../../../../shared/api-client/generated/meals-requests';
import { fetchShareGroups, type ShareGroup, type Visibility } from '../../../../shared/sharing/sharing';
import { MEALS, type Meal, type MealEntry } from '../../domain/meals';
import type { MealData, MealsGateway } from '../../usecase/mealsUseCases';

const toMeal = (m: string): Meal => ((MEALS as readonly string[]).includes(m) ? (m as Meal) : '');

const toEntry = (d: EntryDTO): MealEntry => ({
  id: d.id,
  date: d.date,
  meal: toMeal(d.meal),
  dish: d.dish,
  link: d.link,
  ingredients: d.ingredients ?? [],
  visibility: d.visibility,
  mine: d.mine,
});

const toBody = (d: MealData, visibility?: Visibility): EntryInputDTO => ({ ...d, visibility });

/** MealsGateway backed by the REST API. */
export class MealsApi implements MealsGateway {
  constructor(private readonly http: HttpClient) {}

  async range(from: string, days: number): Promise<MealEntry[]> {
    const q = new URLSearchParams({ from, days: String(days) });
    return (await this.http.get<EntryDTO[]>(`/api/meals?${q}`)).map(toEntry);
  }

  async create(data: MealData, visibility?: Visibility): Promise<MealEntry> {
    return toEntry(await this.http.post<EntryDTO>('/api/meals', toBody(data, visibility)));
  }

  async update(id: string, data: MealData, visibility?: Visibility): Promise<MealEntry> {
    return toEntry(
      await this.http.put<EntryDTO>(`/api/meals/${encodeURIComponent(id)}`, toBody(data, visibility)),
    );
  }

  remove(id: string): Promise<void> {
    return this.http.delete(`/api/meals/${encodeURIComponent(id)}`);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return fetchShareGroups(this.http);
  }
}
