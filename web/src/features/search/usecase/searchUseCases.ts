import type { SearchGateway, SearchSettings, Shortcut } from './ports';

/** Application logic of the search feature. */
export class SearchUseCases {
  constructor(private readonly gateway: SearchGateway) {}

  /** Returns the URL to navigate to for a query, or null for blank input. */
  targetFor(query: string): string | null {
    const q = query.trim();
    return q === '' ? null : this.gateway.searchUrl(q);
  }

  settings(): Promise<SearchSettings> {
    return this.gateway.settings();
  }

  setDefaultProvider(id: string): Promise<SearchSettings> {
    return this.gateway.setDefaultProvider(id);
  }

  /** Adds a shortcut; a leading "!" in the keyword is optional. */
  addShortcut(s: Shortcut): Promise<SearchSettings> {
    return this.gateway.addShortcut({ ...s, keyword: s.keyword.trim().replace(/^!/, '').toLowerCase() });
  }

  deleteShortcut(keyword: string): Promise<SearchSettings> {
    return this.gateway.deleteShortcut(keyword);
  }
}
