import AxeBuilder from '@axe-core/playwright';
import { expect, type Page } from '@playwright/test';

/** WCAG 2.0–2.2 A and AA rules (#113). */
const tags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'];

/**
 * Runs axe on the page (or a part of it) and fails with a readable list of
 * violations: rule, impact and the offending elements.
 */
export async function expectAccessible(page: Page, view: string, include?: string): Promise<void> {
  // Fades (hover, focus) would let axe measure colours mid-transition; with
  // reduced motion the app switches transitions off, so every state is final.
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.mouse.move(0, 0);
  let builder = new AxeBuilder({ page }).withTags(tags);
  if (include) builder = builder.include(include);
  const { violations } = await builder.analyze();
  const report = violations.map(
    // At most three elements per rule keep the report readable.
    (v) =>
      `${v.id} (${v.impact}): ${v.help}\n${v.nodes
        .slice(0, 3)
        .map((n) => `    ${n.target.join(' ')} ${n.failureSummary?.split('\n')[1] ?? ''}`)
        .join('\n')}`,
  );
  // Soft: one run lists the violations of every view.
  expect.soft(report, `accessibility violations in ${view}`).toEqual([]);
}
