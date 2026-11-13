import type { Tab, WidgetConfig } from './dashboard';

/**
 * The tab bar is only worth its space with more than one tab; in edit
 * mode it stays so a tab can be added (#167).
 */
export function showTabBar(tabCount: number, editing: boolean): boolean {
  return tabCount > 1 || editing;
}

/**
 * Configuration of the first widget of a kind, looking at the active tab
 * first, then at the others in their order; null if there is none.
 */
export function firstWidgetConfig(tabs: readonly Tab[], activeId: string, kind: string): WidgetConfig | null {
  const ordered = [...tabs.filter((t) => t.id === activeId), ...tabs.filter((t) => t.id !== activeId)];
  for (const tab of ordered) {
    const w = [...tab.widgets].sort((a, b) => a.y - b.y || a.x - b.x).find((x) => x.kind === kind);
    if (w) return w.config;
  }
  return null;
}
