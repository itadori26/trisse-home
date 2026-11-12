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
  await expect(page.getByText('DuckDuckGo', { exact: true })).toBeVisible({ timeout: 15_000 }); // lazy chunk + first fetch
  await expect(page.getByText(/CPU \(\d+ cores\)/)).toBeVisible({ timeout: 15_000 });
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
  await expect(page.getByRole('option', { name: /Docker documentation/ })).toHaveAttribute(
    'aria-selected',
    'true',
  );
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

test('the UI follows a German browser language', async ({ browser }) => {
  const context = await browser.newContext({ locale: 'de-DE' });
  const page = await context.newPage();
  await page.goto('/');
  await page.getByLabel('Passwort').fill(password);
  await page.getByRole('button', { name: 'Anmelden', exact: true }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Einstellungen' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Bearbeiten', exact: true })).toBeVisible();
  await expect(page.locator('html')).toHaveAttribute('lang', 'de');
  await context.close();
});

test('a guest sees only the shared tab and cannot edit', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.getByRole('button', { name: 'Settings', exact: true }).click();
  await page.getByRole('button', { name: 'Guests', exact: true }).click();
  const form = page.locator('form.new');
  await form.getByLabel('Guest name').fill('anna');
  await form.getByLabel('Password (at least 12 characters)').fill('guest-password-1');
  await form.getByLabel('Repeat password').fill('guest-password-1');
  await form.getByLabel('Home', { exact: true }).check();
  await page.getByRole('button', { name: 'Create guest' }).click();
  await expect(page.getByText('anna', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Close' }).click();
  await page.getByRole('button', { name: 'Log out' }).click();

  await page.getByRole('button', { name: 'Sign in as guest' }).click();
  await page.getByLabel('Guest name').fill('anna');
  await page.getByLabel('Password').fill('guest-password-1');
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Home', exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Server', exact: true })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Edit', exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Settings', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Guests', exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Close' }).click();
  await page.getByRole('button', { name: 'Log out' }).click();
  await expect(page.getByRole('button', { name: 'Log in' })).toBeVisible();
});

test('every widget kind loads its lazy chunk (#116)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  const status = await (await page.request.get('/api/auth/status')).json();
  const headers = { 'X-CSRF-Token': status.csrfToken as string };
  const { widgetKinds } = await (await page.request.get('/api/modules')).json();
  const tab = await (
    await page.request.post('/api/dashboard/tabs', { headers, data: { name: 'All widgets' } })
  ).json();
  for (const [i, kind] of (widgetKinds as string[]).entries()) {
    const res = await page.request.post(`/api/dashboard/tabs/${tab.id}/widgets`, {
      headers,
      data: { kind, x: (i % 3) * 4, y: Math.floor(i / 3) * 4, w: 4, h: 4 },
    });
    expect(res.ok(), kind).toBeTruthy();
  }

  await page.reload();
  await page.getByRole('button', { name: 'All widgets' }).click();
  await expect(page.locator('[aria-busy="true"]')).toHaveCount(0, { timeout: 15_000 });
  await expect(page.getByText('This widget could not be loaded.')).toHaveCount(0);
});

test('the widget picker is searchable and works with the keyboard (#104)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.getByRole('button', { name: 'Edit', exact: true }).click();
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Picker');
  await page.getByRole('button', { name: 'Create' }).click();

  await page.getByRole('button', { name: '+ Add widget' }).click();
  const search = page.getByRole('searchbox', { name: 'Search widgets…' });
  await expect(search).toBeFocused();
  await expect(page.getByRole('heading', { name: 'Everyday' })).toBeVisible();
  await search.fill('vpn');
  await expect(page.getByRole('status')).toHaveText('1 widgets');
  await expect(page.getByRole('button', { name: /^WireGuard/ })).toBeVisible();
  await search.fill('zzz');
  await expect(page.getByText('No widget matches “zzz”.')).toBeVisible();

  await search.fill('uhr');
  await search.press('ArrowDown');
  await expect(page.getByRole('button', { name: /^Clock/ })).toBeFocused();
  await page.keyboard.press('Enter');
  // The clock has settings; they open right after adding.
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await page.getByRole('button', { name: 'Done', exact: true }).click();
  await expect(page.locator('time')).toBeVisible();
});

test('keyboard shortcuts switch tabs, toggle edit mode and show an overview (#95)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.locator('body').click();

  await page.keyboard.press('?');
  await expect(page.getByRole('dialog', { name: 'Keyboard shortcuts' })).toBeVisible();
  await expect(page.getByText('Switch to tab 1 to 9')).toBeVisible();
  await page.keyboard.press('Escape');

  await page.keyboard.press('2');
  await expect(page.getByRole('button', { name: 'Server', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  );
  await page.keyboard.press('1');
  await expect(page.getByRole('button', { name: 'Home', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  );

  await page.keyboard.press('e');
  await expect(page.getByRole('button', { name: 'Done', exact: true })).toBeVisible();
  await page.keyboard.press('e');
  await expect(page.getByRole('button', { name: 'Edit', exact: true })).toBeVisible();

  await page.keyboard.press('/');
  await expect(page.getByRole('dialog', { name: /Quick search/ })).toBeVisible();
  // Typing in the search field must not trigger shortcuts.
  await page.keyboard.type('e2');
  await expect(page.getByRole('button', { name: 'Edit', exact: true })).toBeVisible();
});

test('removed widgets come back with undo, and redo works (#93)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.getByRole('button', { name: 'Edit', exact: true }).click();
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Undo');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('button', { name: /^Clock/ }).click();
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.locator('time')).toBeVisible();

  await page.getByRole('button', { name: 'Remove widget' }).click();
  await expect(page.locator('time')).toHaveCount(0);
  await page.getByRole('status').filter({ hasText: 'removed' }).getByRole('button', { name: 'Undo' }).click();
  await expect(page.locator('time')).toBeVisible();

  // History now: add, configure (the clock's settings were saved). Undo
  // both with the keyboard: the widget is gone again.
  await page.locator('body').click();
  await page.keyboard.press('Control+z');
  await page.keyboard.press('Control+z');
  await expect(page.locator('time')).toHaveCount(0);
  await page.keyboard.press('Control+Shift+z');
  await expect(page.locator('time')).toBeVisible();

  // The restored widget survives a reload.
  await page.reload();
  await page.getByRole('button', { name: 'Undo', exact: true }).click();
  await expect(page.locator('time')).toBeVisible();
});

test('quick capture creates items from the search bar and the palette (#105, #67)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  const search = page.getByRole('searchbox', { name: 'Search' });

  await search.fill('t Milch kaufen morgen');
  await expect(page.getByText('Enter: create to-do “Milch kaufen” (Tomorrow)')).toBeVisible();
  await search.press('Enter');
  await expect(page.getByText('To-do “Milch kaufen” created')).toBeVisible();
  const todos = await (await page.request.get('/api/todos')).json();
  expect(JSON.stringify(todos)).toContain('Milch kaufen');

  await search.fill('n Idee Fahrradtour');
  await search.press('Enter');
  await expect(page.getByText('Note “Idee Fahrradtour” created')).toBeVisible();
  await page.getByRole('status').filter({ hasText: 'created' }).getByRole('button', { name: 'Undo' }).click();

  await search.fill('b example.org Beispiel');
  await search.press('Enter');
  await expect(page.getByText('Bookmark “Beispiel” saved')).toBeVisible();
  await expect(page.getByRole('status').getByRole('link', { name: 'Open' })).toHaveAttribute(
    'href',
    'https://example.org/',
  );

  await page.keyboard.press('Control+k');
  const palette = page.getByRole('combobox');
  await palette.fill('todo: Steuer heute');
  await expect(page.getByText('create to-do “Steuer” (Today)')).toBeVisible();
  await palette.press('Enter');
  await expect(page.getByText('To-do “Steuer” created')).toBeVisible();
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
