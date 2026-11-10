/** A selectable search provider. */
export interface Provider {
  id: string;
  name: string;
}

/** A custom shortcut such as !gh → GitHub search. */
export interface Shortcut {
  keyword: string;
  name: string;
  urlTemplate: string;
}

/** The search configuration. */
export interface SearchSettings {
  defaultProvider: string;
  providers: Provider[];
  shortcuts: Shortcut[];
}

/** Output port: talks to the search backend. */
export interface SearchGateway {
  settings(): Promise<SearchSettings>;
  setDefaultProvider(id: string): Promise<SearchSettings>;
  addShortcut(s: Shortcut): Promise<SearchSettings>;
  deleteShortcut(keyword: string): Promise<SearchSettings>;
  /** URL that resolves a query (the backend redirects to the provider). */
  searchUrl(query: string): string;
}
