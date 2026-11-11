import type { MessageKey } from '../../../shared/i18n';

/** One step of the guided tour. */
export interface TourStep {
  id: string;
  title: MessageKey;
  body: MessageKey;
  /** Value of the data-tour attribute of the element to highlight; none = centered. */
  target?: string;
}

/** All steps; steps whose target is missing on the page are skipped by the UI. */
export const tourSteps: TourStep[] = [
  {
    id: 'welcome',
    title: 'tour.welcome.title',
    body: 'tour.welcome.body',
  },
  {
    id: 'tabs',
    target: 'tabs',
    title: 'tour.tabs.title',
    body: 'tour.tabs.body',
  },
  {
    id: 'search',
    target: 'search',
    title: 'tour.search.title',
    body: 'tour.search.body',
  },
  {
    id: 'edit',
    target: 'edit',
    title: 'tour.edit.title',
    body: 'tour.edit.body',
  },
  {
    id: 'settings',
    target: 'settings',
    title: 'tour.settings.title',
    body: 'tour.settings.body',
  },
  {
    id: 'server',
    target: 'tab-Server',
    title: 'tour.server.title',
    body: 'tour.server.body',
  },
  {
    id: 'done',
    title: 'tour.done.title',
    body: 'tour.done.body',
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
