import { createHmac } from 'node:crypto';
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

  // The guided tour starts automatically after the first login.
  const tour = page.getByRole('dialog', { name: 'Welcome to trisse-home 👋' });
  await expect(tour).toBeVisible();
  for (const title of ['Tabs', 'Search', 'Edit mode', 'Settings', 'Server overview']) {
    await page.getByRole('button', { name: 'Next', exact: true }).click();
    await expect(page.getByRole('dialog', { name: title })).toBeVisible();
  }
  await page.getByRole('button', { name: 'Next', exact: true }).click();
  await page.getByRole('button', { name: "Let's go" }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  // After the first setup the dashboard is filled with sample widgets and data.
  await expect(page.getByRole('search')).toBeVisible();
  await expect(page.locator('time')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Welcome to trisse-home 👋' })).toBeVisible();
  await expect(page.getByText('Try edit mode: drag and resize a widget')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Docker documentation' })).toBeVisible();
  await page.getByRole('button', { name: 'Server', exact: true }).click();
  await expect(page.getByText('DuckDuckGo', { exact: true })).toBeVisible();
  await expect(page.getByText(/CPU \(\d+ cores\)/)).toBeVisible();
  await page.getByRole('button', { name: 'Home', exact: true }).click();

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
  // A completed tour is not shown again.
  await page.waitForTimeout(1000);
  await expect(page.getByRole('dialog')).toHaveCount(0);
});

test('edit mode adds a tab and a widget', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.getByRole('button', { name: 'Edit', exact: true }).click();
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Study');
  await page.getByRole('button', { name: 'Create' }).click();
  await expect(page.getByRole('button', { name: 'Study' })).toHaveAttribute('aria-current', 'page');

  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('button', { name: /^Clock/ }).click();
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await page.getByRole('button', { name: 'Done', exact: true }).click();
  await expect(page.locator('time')).toBeVisible();

  await page.reload();
  await expect(page.getByRole('button', { name: 'Study' })).toHaveAttribute('aria-current', 'page');
  await expect(page.locator('time')).toBeVisible();
});

test('widgets can be moved and resized with the keyboard', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Study' }).click();
  await page.getByRole('button', { name: 'Edit', exact: true }).click();

  const clock = page.getByRole('group', { name: /^Clock, column 1, row 1/ });
  await clock.focus();
  await page.keyboard.press('ArrowRight');
  await expect(page.getByText(/^Clock moved: column 2, row 1/)).toBeAttached();
  await page.keyboard.press('Shift+ArrowRight');
  await expect(page.getByText(/^Clock resized: column 2, row 1, \d+ × \d+ cells/)).toBeAttached();
  await expect(page.getByRole('group', { name: /^Clock, column 2, row 1/ })).toBeFocused();
  await page.getByRole('button', { name: 'Done', exact: true }).click();

  await page.reload();
  await page.getByRole('button', { name: 'Study' }).click();
  await page.getByRole('button', { name: 'Edit', exact: true }).click();
  await expect(page.getByRole('group', { name: /^Clock, column 2, row 1/ })).toBeVisible();
});

/** RFC 6238 TOTP for the test (SHA1, 6 digits, 30 s). */
function totp(base32Secret: string, now = Date.now()): string {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  let bits = '';
  for (const ch of base32Secret.replace(/\s/g, '').toUpperCase())
    bits += alphabet.indexOf(ch).toString(2).padStart(5, '0');
  const key = Buffer.from((bits.match(/.{8}/g) ?? []).map((b) => parseInt(b, 2)));
  const counter = Buffer.alloc(8);
  counter.writeBigUInt64BE(BigInt(Math.floor(now / 1000 / 30)));
  const mac = createHmac('sha1', key).update(counter).digest();
  const offset = (mac[mac.length - 1] ?? 0) & 0x0f;
  return String((mac.readUInt32BE(offset) & 0x7fffffff) % 1_000_000).padStart(6, '0');
}

test('quick search finds bookmarks with the keyboard', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: /Search bookmarks/ });
  await expect(box).toBeFocused();
  await box.fill('docker docs');
  await expect(page.getByRole('option', { name: /Docker documentation/ })).toHaveAttribute('aria-selected', 'true');
  await box.fill('nothing-matches-this');
  await expect(page.getByText(/No matches/)).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(box).toBeHidden();

  await page.keyboard.press('/');
  await expect(page.getByRole('combobox', { name: /Search bookmarks/ })).toBeFocused();
});

test('installable app shows the last data offline and never caches auth', async ({ page, context }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await expect(page.locator('link[rel="manifest"]')).toHaveAttribute('href', '/manifest.webmanifest');
  await page.evaluate(() => navigator.serviceWorker.ready);
  await page.reload(); // the worker controls the page from now on
  await expect(page.getByTestId('dashboard')).toBeVisible();

  const cached = await page.evaluate(async () => {
    const cache = await caches.open('trisse-data-v1');
    return (await cache.keys()).map((r) => new URL(r.url).pathname);
  });
  expect(cached).toContain('/api/dashboard');
  expect(cached.some((p) => p.startsWith('/api/auth'))).toBe(false);

  await context.setOffline(true);
  await page.reload();
  await expect(page.getByText(/You are offline/)).toBeVisible();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await context.setOffline(false);
});

test('two-factor authentication can be enabled and is required at login', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.getByRole('button', { name: 'Settings', exact: true }).click();
  await page.getByRole('button', { name: 'Account', exact: true }).click();
  await page.getByRole('button', { name: 'Set up two-factor authentication' }).click();
  await expect(page.getByAltText('QR code for the authenticator app')).toBeVisible();
  const secret = (await page.getByRole('dialog').locator('code').textContent()) ?? '';
  await page.getByLabel('Code from the app').fill(totp(secret));
  await page.getByRole('button', { name: 'Turn on' }).click();
  await expect(page.getByText('Two-factor authentication is on.')).toBeVisible();
  await page.getByRole('button', { name: 'I saved them' }).click();
  await page.getByRole('button', { name: 'Close' }).click();

  await page.getByRole('button', { name: 'Log out' }).click();
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByLabel('Authentication code (or recovery code)')).toBeVisible();
  // Wait for the next time step: the code used for enabling cannot be replayed.
  const step = Math.floor(Date.now() / 30000);
  await expect.poll(() => Math.floor(Date.now() / 30000), { timeout: 35_000 }).toBeGreaterThan(step);
  await page.getByLabel('Authentication code (or recovery code)').fill(totp(secret));
  await page.getByRole('button', { name: 'Verify' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
});
