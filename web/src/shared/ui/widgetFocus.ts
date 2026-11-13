/**
 * A small channel between widgets and the dashboard: a widget asks to show
 * another widget (e.g. the briefing opens the to-do widget, #52) without
 * knowing the dashboard. The dashboard switches to a tab with a matching
 * widget and scrolls it into view.
 */
const EVENT = 'focus-widget';
const bus = new EventTarget();

/** Asks the dashboard to show a widget of one of the kinds (first match wins). */
export function requestWidgetFocus(kinds: readonly string[]): void {
  bus.dispatchEvent(new CustomEvent<string[]>(EVENT, { detail: [...kinds] }));
}

/** Subscribes to focus requests; returns the unsubscribe function. */
export function onWidgetFocusRequest(handler: (kinds: string[]) => void): () => void {
  const listener = (e: Event) => handler((e as CustomEvent<string[]>).detail);
  bus.addEventListener(EVENT, listener);
  return () => bus.removeEventListener(EVENT, listener);
}

/**
 * Picks the widget to show: a matching widget on the active tab first,
 * otherwise the first match on another tab, kinds in order of preference.
 */
export function pickWidget<W extends { id: string; kind: string }, T extends { id: string; widgets: W[] }>(
  tabs: readonly T[],
  activeId: string,
  kinds: readonly string[],
): { tabId: string; widgetId: string } | null {
  const ordered = [...tabs].sort((a, b) => (a.id === activeId ? -1 : b.id === activeId ? 1 : 0));
  for (const kind of kinds) {
    for (const tab of ordered) {
      const w = tab.widgets.find((x) => x.kind === kind);
      if (w) return { tabId: tab.id, widgetId: w.id };
    }
  }
  return null;
}
