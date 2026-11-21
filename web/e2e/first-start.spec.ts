import { expect, test } from '@playwright/test';
import { password, setupToken } from './support';

// The first start with the default SAMPLE_DATA=true (#670), on a server of
// its own (see playwright.config.ts): the template choice comes before the
// tour, and a template replaces the example tabs with a zoned tab.
test('first start offers the templates and Family replaces the example dashboard', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Setup token').fill(setupToken);
  await page.getByLabel('Password (at least 12 characters)').fill(password);
  await page.getByLabel('Repeat password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  const picker = page.getByRole('dialog', { name: 'What do you use Waylight for?' });
  await expect(picker).toBeVisible();
  await expect(picker.getByText('it replaces the example dashboard')).toBeVisible();
  // The tour waits for the choice.
  await expect(page.getByRole('dialog', { name: 'Welcome to Waylight 👋' })).toHaveCount(0);

  await picker.getByText('Family', { exact: true }).click();
  await picker.getByRole('button', { name: 'Use this template' }).click();
  // The weather module is off on the test server; the location can wait.
  await page
    .getByRole('dialog', { name: 'Where are you?' })
    .getByRole('button', { name: 'Set it up later' })
    .click();

  const tour = page.getByRole('dialog', { name: 'Welcome to Waylight 👋' });
  await expect(tour).toBeVisible();
  await tour.getByRole('button', { name: 'Skip tour' }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);

  // Only the template's tab is left (a single tab shows no tab bar), laid
  // out in zones; the owner sees System.
  await expect(page.getByRole('button', { name: 'Server', exact: true })).toHaveCount(0);
  for (const [zone, heading] of [
    ['today', 'Today'],
    ['mine', 'My things'],
    ['system', 'System'],
  ] as const) {
    await expect(
      page.locator(`section[data-zone="${zone}"]`).getByRole('heading', { name: heading }),
    ).toBeVisible();
  }

  const dashboard = (await (await page.request.get('/api/dashboard')).json()) as {
    tabs: {
      name: string;
      zoned: boolean;
      sample?: boolean;
      widgets: { kind: string; size?: string; zone?: string }[];
    }[];
  };
  expect(dashboard.tabs.map((t) => t.name)).toEqual(['Family']);
  expect(dashboard.tabs[0]?.zoned).toBe(true);
  expect(dashboard.tabs[0]?.widgets.find((w) => w.kind === 'services')).toMatchObject({
    size: 'strip',
    zone: 'system',
  });
  // The example data of the modules stays.
  expect(((await (await page.request.get('/api/todos')).json()) as unknown[]).length).toBeGreaterThan(0);
});
