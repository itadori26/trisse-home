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

/** Quick capture prefixes of the search field, e.g. "t" for "t Milch kaufen" (#221). */
export interface CapturePrefixes {
  todo: string;
  note: string;
  bookmark: string;
}

/** Options of the search field (#221), formerly options of the search widget. */
export interface SearchBarOptions {
  openInNewTab: boolean;
  capturePrefixes: CapturePrefixes;
}

/** The search configuration. */
export interface SearchSettings {
  defaultProvider: string;
  providers: Provider[];
  shortcuts: Shortcut[];
  bar: SearchBarOptions;
}

/** Output port: talks to the search backend. */
export interface SearchGateway {
  settings(): Promise<SearchSettings>;
  setDefaultProvider(id: string): Promise<SearchSettings>;
  setBar(options: SearchBarOptions): Promise<SearchSettings>;
  addShortcut(s: Shortcut): Promise<SearchSettings>;
  deleteShortcut(keyword: string): Promise<SearchSettings>;
  /** URL that resolves a query (the backend redirects to the provider). */
  searchUrl(query: string): string;
}
