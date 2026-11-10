import { expect, test } from '@playwright/test';
import { setupToken } from '../playwright.config';

const password = 'smoke-test-password-123';

test('setup, logout and login lead to the dashboard', async ({ page }) => {
  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Welcome to trisse-home' })).toBeVisible();
  await page.getByLabel('Setup token').fill(setupToken);
  await page.getByLabel('Password (at least 12 characters)').fill(password);
  await page.getByLabel('Repeat password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  // The first tab is seeded with a search bar and a clock.
  await expect(page.getByRole('search')).toBeVisible();
  await expect(page.locator('time')).toBeVisible();

  await page.getByRole('button', { name: 'Log out' }).click();
  await expect(page.getByRole('button', { name: 'Log in' })).toBeVisible();

  await page.getByLabel('Password').fill('wrong-password-123');
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByRole('alert')).toContainText('incorrect');

  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.reload();
  await expect(page.getByTestId('dashboard')).toBeVisible();
});

test('edit mode adds a tab and a widget', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.getByRole('button', { name: 'Edit' }).click();
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Server');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByRole('button', { name: 'Server' })).toHaveAttribute('aria-current', 'page');

  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('button', { name: /^Clock/ }).click();
  await page.getByRole('button', { name: 'Save' }).click();
  await page.getByRole('button', { name: 'Done' }).click();
  await expect(page.locator('time')).toBeVisible();

  await page.reload();
  await expect(page.getByRole('button', { name: 'Server' })).toHaveAttribute('aria-current', 'page');
  await expect(page.locator('time')).toBeVisible();
});
