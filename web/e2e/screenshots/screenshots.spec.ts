import { expect, test, type Page } from '@playwright/test';

/** Where the images go; referenced from the README. */
const outDir = '../docs/screenshots';

const variants = [
  { name: 'desktop-light', scheme: 'light', viewport: { width: 1440, height: 900 } },
  { name: 'desktop-dark', scheme: 'dark', viewport: { width: 1440, height: 900 } },
  { name: 'mobile-light', scheme: 'light', viewport: { width: 390, height: 844 } },
  { name: 'mobile-dark', scheme: 'dark', viewport: { width: 390, height: 844 } },
] as const;

async function openDashboard(page: Page) {
  await page.goto('/');
  // The guided tour starts for the demo account: skip it once (it is stored).
  const skip = page.getByRole('button', { name: 'Skip tour' });
  await skip.click({ timeout: 10_000 }).catch(() => undefined);
  await expect(page.locator('.tour')).toHaveCount(0);
  await expect(page.getByRole('search')).toBeVisible({ timeout: 15_000 });
  // The demo banner is not part of the product screenshots.
  await page.addStyleTag({ content: '.demo-banner { display: none !important; }' });
  await settle(page);
}

/** Waits until widgets have loaded their data. */
async function settle(page: Page) {
  await expect(page.locator('[aria-busy="true"]')).toHaveCount(0, { timeout: 15_000 });
  await page.waitForTimeout(1500);
}

for (const v of variants) {
  test(`dashboard ${v.name}`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: v.scheme, reducedMotion: 'reduce' });
    await page.setViewportSize(v.viewport);
    await openDashboard(page);
    await page.screenshot({
      path: `${outDir}/home-${v.name}.png`,
      fullPage: v.name.startsWith('mobile') ? false : true,
    });
    if (v.name.startsWith('desktop')) {
      await page.getByRole('button', { name: 'Server', exact: true }).click();
      await expect(page.getByText(/CPU \(\d+ cores\)/)).toBeVisible({ timeout: 15_000 });
      await settle(page);
      await page.screenshot({ path: `${outDir}/server-${v.name}.png`, fullPage: true });
    }
  });
}
