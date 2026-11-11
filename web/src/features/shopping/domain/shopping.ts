/** One entry of the shopping list. */
export interface ShoppingItem {
  id: string;
  summary: string;
  done: boolean;
}

/** Connection settings as shown in the UI (the password is never returned). */
export interface ShoppingSettings {
  configured: boolean;
  listUrl: string;
  username: string;
  hasPassword: boolean;
}

/** Prefill values derived from another CalDAV configuration. */
export interface ShoppingSuggestion {
  listUrl: string;
  username: string;
}

/**
 * Suggests a task list URL next to a calendar collection, e.g.
 * …/calendars/me/personal/ → …/calendars/me/shopping/.
 */
export function suggestListUrl(calendarUrl: string): string {
  const m = /^(.*\/)[^/]+\/?$/.exec(calendarUrl.trim());
  return m ? `${m[1]}shopping/` : '';
}

/** Applies a done change without mutating the input (optimistic update). */
export function withDone(items: ShoppingItem[], id: string, done: boolean): ShoppingItem[] {
  return items.map((i) => (i.id === id ? { ...i, done } : i));
}
