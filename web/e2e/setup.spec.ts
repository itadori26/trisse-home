import { expect, test } from '@playwright/test';
import { account, password, setupToken } from './support';

test('setup, logout and login lead to the dashboard', async ({ page }) => {
  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Welcome to Waylight' })).toBeVisible();
  // The setup asks no start-page profile (#655): the sample data includes the Server tab.
  await expect(page.getByRole('radio')).toHaveCount(0);
  await page.getByLabel('Setup token').fill(setupToken);
  await page.getByLabel('Password (at least 12 characters)').fill(password);
  await page.getByLabel('Repeat password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  // With sample data the template choice comes first (#670). The other
  // specs build on the example dashboard, so it is kept here;
  // first-start.spec.ts chooses a template on a server of its own.
  const picker = page.getByRole('dialog', { name: 'What do you use Waylight for?' });
  await expect(picker).toBeVisible();
  await picker.getByRole('button', { name: 'Keep the example dashboard' }).click();

  // The guided tour follows the template choice.
  const tour = page.getByRole('dialog', { name: 'Welcome to Waylight 👋' });
  await expect(tour).toBeVisible();
  for (const title of ['Tabs', 'Search', 'Edit mode', 'Settings', 'Server overview']) {
    await page.getByRole('button', { name: 'Next', exact: true }).click();
    await expect(page.getByRole('dialog', { name: title })).toBeVisible();
  }
  await page.getByRole('button', { name: 'Next', exact: true }).click();
  await page.getByRole('button', { name: "Let's go" }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  // After the first setup the dashboard is filled with sample widgets and data.
  await expect(page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' })).toBeVisible();
  await expect(page.locator('time')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Welcome to Waylight 👋' })).toBeVisible();
  await expect(page.getByText('Try edit mode: drag and resize a widget')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Docker documentation' })).toBeVisible();
  await page.getByRole('button', { name: 'Server', exact: true }).click();
  await expect(page.getByText('DuckDuckGo', { exact: true })).toBeVisible({ timeout: 15_000 }); // lazy chunk + first fetch
  await expect(page.getByText(/CPU \(\d+ cores\)/)).toBeVisible({ timeout: 15_000 });
  await page.getByRole('button', { name: 'Home', exact: true }).click();

  await account(page, 'Log out');
  await expect(page.getByRole('button', { name: 'Log in' })).toBeVisible();

  await page.getByLabel('Password').fill('wrong-password-123');
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByRole('alert')).toContainText('incorrect');

  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.reload();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  // A completed tour is not shown again.
  await page.waitForTimeout(1000);
  await expect(page.getByRole('dialog')).toHaveCount(0);
});
