/** One step of the guided tour. */
export interface TourStep {
  id: string;
  title: string;
  body: string;
  /** Value of the data-tour attribute of the element to highlight; none = centered. */
  target?: string;
}

/** All steps; steps whose target is missing on the page are skipped by the UI. */
export const tourSteps: TourStep[] = [
  {
    id: 'welcome',
    title: 'Welcome to trisse-home 👋',
    body: 'Your personal start page: widgets for everyday life and your home server. This short tour shows you around — it takes less than a minute.',
  },
  {
    id: 'tabs',
    target: 'tabs',
    title: 'Tabs',
    body: 'Group widgets on tabs, e.g. "Home" for every day and "Server" for your homelab. The selected tab is remembered.',
  },
  {
    id: 'search',
    target: 'search',
    title: 'Search',
    body: 'Searches with DuckDuckGo by default. Bangs like !w or !g work, and you can define your own shortcuts in the settings.',
  },
  {
    id: 'edit',
    target: 'edit',
    title: 'Edit mode',
    body: 'Switch to edit mode to drag widgets by their title bar, resize them at the bottom-right corner, add new widgets and tabs, and configure a widget with ⚙.',
  },
  {
    id: 'settings',
    target: 'settings',
    title: 'Settings',
    body: 'Theme and background, search engine, service tiles, RSS feeds, bookmark folders, your CalDAV calendar and AdGuard Home — and your password.',
  },
  {
    id: 'server',
    target: 'tab-Server',
    title: 'Server overview',
    body: 'The "Server" tab shows host metrics, Docker containers and the status of your services.',
  },
  {
    id: 'done',
    title: "You're all set!",
    body: 'The sample data is just a starting point — change or delete whatever you like. You can restart this tour any time under Settings → Account.',
  },
];

/** Keeps only steps that are centered or whose target exists. */
export function availableSteps(
  steps: readonly TourStep[],
  hasTarget: (target: string) => boolean,
): TourStep[] {
  return steps.filter((s) => !s.target || hasTarget(s.target));
}

/** Rectangle in viewport coordinates. */
export interface Box {
  top: number;
  left: number;
  width: number;
  height: number;
}

/**
 * Places the popover below the target if there is room, otherwise above it,
 * horizontally aligned with the target and kept inside the viewport.
 * Without target the popover is centered.
 */
export function popoverPosition(
  target: Box | null,
  popover: { width: number; height: number },
  viewport: { width: number; height: number },
  gap = 12,
  margin = 12,
): { top: number; left: number } {
  const clampX = (x: number) =>
    Math.min(Math.max(x, margin), Math.max(margin, viewport.width - popover.width - margin));
  const clampY = (y: number) =>
    Math.min(Math.max(y, margin), Math.max(margin, viewport.height - popover.height - margin));
  if (!target) {
    return {
      top: clampY((viewport.height - popover.height) / 2),
      left: clampX((viewport.width - popover.width) / 2),
    };
  }
  const below = target.top + target.height + gap;
  const above = target.top - gap - popover.height;
  const top = below + popover.height + margin <= viewport.height || above < margin ? below : above;
  return { top: clampY(top), left: clampX(target.left + target.width / 2 - popover.width / 2) };
}
