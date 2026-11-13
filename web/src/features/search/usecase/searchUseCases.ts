import type { SearchBarOptions, SearchGateway, SearchSettings, Shortcut } from './ports';

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

  /** Saves the options of the search field; prefixes are trimmed and lower-cased (#221). */
  setBar(options: SearchBarOptions): Promise<SearchSettings> {
    const p = options.capturePrefixes;
    const norm = (s: string) => s.trim().toLowerCase();
    return this.gateway.setBar({
      openInNewTab: options.openInNewTab,
      capturePrefixes: { todo: norm(p.todo), note: norm(p.note), bookmark: norm(p.bookmark) },
    });
  }

  /** Adds a shortcut; a leading "!" in the keyword is optional. */
  addShortcut(s: Shortcut): Promise<SearchSettings> {
    return this.gateway.addShortcut({ ...s, keyword: s.keyword.trim().replace(/^!/, '').toLowerCase() });
  }

  deleteShortcut(keyword: string): Promise<SearchSettings> {
    return this.gateway.deleteShortcut(keyword);
  }
}
