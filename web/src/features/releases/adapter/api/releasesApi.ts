import type { HttpClient } from '../../../../shared/api-client/http';
import type { ArtistDTO, ListDTO } from '../../../../shared/api-client/generated/releases';
import type { FollowDTO } from '../../../../shared/api-client/generated/releases-requests';
import type { Artist, ReleaseList } from '../../domain/releases';
import type { ReleasesGateway } from '../../usecase/releasesUseCases';

/** ReleasesGateway backed by the REST API. */
export class ReleasesApi implements ReleasesGateway {
  constructor(private readonly http: HttpClient) {}

  async search(query: string): Promise<Artist[]> {
    return (await this.http.get<ArtistDTO[]>(`/api/releases/search?q=${encodeURIComponent(query)}`)).map(
      (a) => ({ ...a }),
    );
  }
  async followed(): Promise<Artist[]> {
    return (await this.http.get<ArtistDTO[]>('/api/releases/artists')).map((a) => ({ ...a }));
  }
  follow(artist: Pick<Artist, 'id' | 'name'>): Promise<void> {
    const body: FollowDTO = { id: artist.id, name: artist.name };
    return this.http.post('/api/releases/artists', body);
  }
  unfollow(id: string): Promise<void> {
    return this.http.delete(`/api/releases/artists/${encodeURIComponent(id)}`);
  }
  async list(weeks: number): Promise<ReleaseList> {
    const d = await this.http.get<ListDTO>(`/api/releases?weeks=${weeks}`);
    return { releases: d.releases, error: d.error };
  }
}
