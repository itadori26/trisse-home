import type { HttpClient } from '../../../../shared/api-client/http';
import type { IdeaDTO } from '../../../../shared/api-client/generated/gifts';
import type { IdeaInputDTO } from '../../../../shared/api-client/generated/gifts-requests';
import type { GiftIdea, GiftInput } from '../../domain/gifts';
import type { GiftsGateway } from '../../usecase/giftsUseCases';

const toIdea = (d: IdeaDTO): GiftIdea => ({
  id: d.id,
  person: d.person,
  title: d.title,
  link: d.link,
  note: d.note,
  bought: d.bought,
});

const toBody = (i: GiftInput): IdeaInputDTO => ({ ...i });
const path = (id: string) => `/api/gifts/${encodeURIComponent(id)}`;

/** GiftsGateway backed by the REST API. */
export class GiftsApi implements GiftsGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<GiftIdea[]> {
    return (await this.http.get<IdeaDTO[]>('/api/gifts')).map(toIdea);
  }

  async create(input: GiftInput): Promise<GiftIdea> {
    return toIdea(await this.http.post<IdeaDTO>('/api/gifts', toBody(input)));
  }

  async update(id: string, input: GiftInput): Promise<GiftIdea> {
    return toIdea(await this.http.put<IdeaDTO>(path(id), toBody(input)));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }
}
