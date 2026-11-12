import type { Artist, ReleaseList } from '../domain/releases';

/** Output port: new-releases backend. */
export interface ReleasesGateway {
  search(query: string): Promise<Artist[]>;
  followed(): Promise<Artist[]>;
  follow(artist: Pick<Artist, 'id' | 'name'>): Promise<void>;
  unfollow(id: string): Promise<void>;
  list(weeks: number): Promise<ReleaseList>;
}

/** Application logic of the new-releases feature (#63). */
export class ReleasesUseCases {
  constructor(private readonly gateway: ReleasesGateway) {}

  /** Searches MusicBrainz; queries shorter than two letters return nothing. */
  search(query: string): Promise<Artist[]> {
    const q = query.trim();
    return q.length < 2 ? Promise.resolve([]) : this.gateway.search(q);
  }

  followed(): Promise<Artist[]> {
    return this.gateway.followed();
  }

  follow(artist: Pick<Artist, 'id' | 'name'>): Promise<void> {
    return this.gateway.follow(artist);
  }

  unfollow(id: string): Promise<void> {
    return this.gateway.unfollow(id);
  }

  list(weeks: number): Promise<ReleaseList> {
    return this.gateway.list(weeks);
  }
}
