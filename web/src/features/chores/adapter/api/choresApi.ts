import type { HttpClient } from '../../../../shared/api-client/http';
import type { ChoreDTO } from '../../../../shared/api-client/generated/chores';
import type { ChoreInputDTO } from '../../../../shared/api-client/generated/chores-requests';
import { fetchShareGroups, type ShareGroup, type Visibility } from '../../../../shared/sharing/sharing';
import type { Chore } from '../../domain/chores';
import type { ChoreData, ChoresGateway } from '../../usecase/choresUseCases';

const toChore = (d: ChoreDTO): Chore => ({
  id: d.id,
  title: d.title,
  everyDays: d.everyDays,
  due: d.due,
  assignees: d.assignees ?? [],
  assignee: d.assignee,
  dueNow: d.dueNow,
  overdue: d.overdue,
  lastDoneAt: d.lastDoneAt ? new Date(d.lastDoneAt) : null,
  lastDoneBy: d.lastDoneBy,
  visibility: d.visibility,
  mine: d.mine,
});

const toBody = (d: ChoreData, visibility?: Visibility): ChoreInputDTO => ({ ...d, visibility });
const path = (id: string) => `/api/chores/${encodeURIComponent(id)}`;

/** ChoresGateway backed by the REST API. */
export class ChoresApi implements ChoresGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<Chore[]> {
    return (await this.http.get<ChoreDTO[]>('/api/chores')).map(toChore);
  }

  async create(data: ChoreData, visibility?: Visibility): Promise<Chore> {
    return toChore(await this.http.post<ChoreDTO>('/api/chores', toBody(data, visibility)));
  }

  async update(id: string, data: ChoreData, visibility?: Visibility): Promise<Chore> {
    return toChore(await this.http.put<ChoreDTO>(path(id), toBody(data, visibility)));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }

  async done(id: string): Promise<Chore> {
    return toChore(await this.http.post<ChoreDTO>(`${path(id)}/done`));
  }

  shareGroups(): Promise<ShareGroup[]> {
    return fetchShareGroups(this.http);
  }
}
