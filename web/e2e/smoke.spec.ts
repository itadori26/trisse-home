import { createHmac } from 'node:crypto';
import { expect, type Page, test } from '@playwright/test';
import { setupToken } from '../playwright.config';

const password = 'smoke-test-password-123';

/** Edit, Done, Settings and Log out live in the account menu of the top bar (#189). */
async function account(page: Page, item: string) {
  await page.getByRole('button', { name: 'Account menu', exact: true }).click();
  await page.getByRole('menuitem', { name: item, exact: true }).click();
}

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

test('edit mode adds a tab and a widget', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Study');
  await page.getByRole('button', { name: 'Create' }).click();
  await expect(page.getByRole('button', { name: 'Study' })).toHaveAttribute('aria-current', 'page');

  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('button', { name: /^Clock/ }).click();
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');
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
  await account(page, 'Edit');

  const clock = page.getByRole('group', { name: /^Clock, column 1, row 1/ });
  await clock.focus();
  await page.keyboard.press('ArrowRight');
  await expect(page.getByText(/^Clock moved: column 2, row 1/)).toBeAttached();
  await page.keyboard.press('Shift+ArrowRight');
  await expect(page.getByText(/^Clock resized: column 2, row 1, \d+ × \d+ cells/)).toBeAttached();
  await expect(page.getByRole('group', { name: /^Clock, column 2, row 1/ })).toBeFocused();
  await account(page, 'Done');

  await page.reload();
  await page.getByRole('button', { name: 'Study' }).click();
  await account(page, 'Edit');
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
  await expect(page.getByRole('button', { name: 'Kontomenü', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Kontomenü', exact: true }).click();
  await expect(page.getByRole('menuitem', { name: 'Bearbeiten', exact: true })).toBeVisible();
  await expect(page.getByRole('menuitem', { name: 'Einstellungen', exact: true })).toBeVisible();
  await expect(page.locator('html')).toHaveAttribute('lang', 'de');
  await context.close();
});

test('settings sections of self-registering features load on demand (#193)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Proxmox', exact: true }).click();
  await expect(page.getByLabel('Proxmox VE URL')).toBeVisible();
});

test('a guest sees only the shared tab and cannot edit', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Guests', exact: true }).click();
  const form = page.locator('form.new');
  await form.getByLabel('Guest name').fill('anna');
  await form.getByLabel('Password (at least 12 characters)').fill('guest-password-1');
  await form.getByLabel('Repeat password').fill('guest-password-1');
  await form.getByLabel('Home', { exact: true }).check();
  await page.getByRole('button', { name: 'Create guest' }).click();
  await expect(page.getByText('anna', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Close' }).click();
  await account(page, 'Log out');

  await page.getByRole('button', { name: 'Sign in as guest' }).click();
  await page.getByLabel('Guest name').fill('anna');
  await page.getByLabel('Password').fill('guest-password-1');
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  // A single shared tab needs no tab bar (#167).
  await expect(page.getByRole('navigation', { name: 'Dashboard tabs' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Server', exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Account menu', exact: true }).click();
  await expect(page.getByRole('menuitem', { name: 'Edit', exact: true })).toHaveCount(0);
  await page.getByRole('menuitem', { name: 'Settings', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Guests', exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Close' }).click();
  await account(page, 'Log out');
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
  await account(page, 'Edit');
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
  await account(page, 'Done');
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
  await expect(page.getByRole('button', { name: '+ Add widget' })).toBeVisible();
  await page.keyboard.press('e');
  await expect(page.getByRole('button', { name: '+ Add widget' })).toHaveCount(0);

  await page.keyboard.press('/');
  await expect(page.getByRole('dialog', { name: /Quick search/ })).toBeVisible();
  // Typing in the search field must not trigger shortcuts.
  await page.keyboard.type('e2');
  await expect(page.getByRole('button', { name: '+ Add widget' })).toHaveCount(0);
});

test('removed widgets come back with undo, and redo works (#93)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Edit');
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
  const search = page.getByRole('searchbox', { name: 'Search', exact: true });

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

test('phones get their own layout that hides widgets independently (#92)', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  await expect(page.locator('time')).toBeVisible();
  const menu = (name: string) => account(page, name);

  await menu('Edit');
  await expect(page.getByText('Phone layout follows the desktop grid.', { exact: false })).toBeVisible();
  await page.getByRole('group', { name: 'Phone layout of “Clock”' }).getByLabel('Hide on phones').check();
  await expect(page.getByRole('button', { name: 'Reset from desktop layout' })).toBeVisible();
  await menu('Done');
  await expect(page.locator('time')).toHaveCount(0);

  await page.reload();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await expect(page.getByText('Try edit mode: drag and resize a widget')).toBeVisible();
  await expect(page.locator('time')).toHaveCount(0);

  // The desktop grid is unaffected.
  await page.setViewportSize({ width: 1280, height: 900 });
  await expect(page.locator('time')).toBeVisible();

  await page.setViewportSize({ width: 390, height: 844 });
  await menu('Edit');
  await page.getByRole('button', { name: 'Reset from desktop layout' }).click();
  await expect(page.getByText('Phone layout follows the desktop grid.', { exact: false })).toBeVisible();
  await menu('Done');
  await expect(page.locator('time')).toBeVisible();
});

test('the daily briefing sums up the day and leads to the source widget (#52)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('briefing');
  await page.getByRole('button', { name: /^Daily briefing/ }).click();
  await account(page, 'Done');

  // "Steuer" was created for today by the quick capture test.
  const toTodos = page.getByRole('button', { name: 'To-dos', exact: true });
  await expect(toTodos).toBeVisible({ timeout: 15_000 });
  await expect(page.locator('.briefing')).toContainText('Steuer');
  await expect(page.locator('.briefing')).not.toContainText('Milch kaufen'); // due tomorrow
  // The waste calendar is not set up yet: that is no outage.
  await expect(page.locator('.briefing')).not.toContainText('Not available right now');

  // The heading scrolls to and highlights the to-do widget.
  await toTodos.click();
  await expect(page.locator('[data-widget-id].flash')).toContainText('Milch kaufen');

  // Hiding the section in the settings removes it from the widget.
  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Daily briefing', exact: true }).click();
  await page.getByRole('checkbox', { name: 'To-dos' }).uncheck();
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByText('Saved.', { exact: true })).toBeVisible();
  await page.reload();
  // Sign-ins of this test run still show up under "Since yesterday evening".
  await expect(page.getByRole('button', { name: 'Since yesterday evening', exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await expect(page.getByRole('button', { name: 'To-dos', exact: true })).toHaveCount(0);
});

test('countdowns count the days and can be edited and deleted (#54)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Dates');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('prüfung');
  await page.getByRole('button', { name: /^Countdowns/ }).click();
  // The widget has settings; they open right after adding.
  await page.getByLabel('Highlight when fewer days are left than').fill('5');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  const inDays = (n: number) => {
    const d = new Date();
    d.setDate(d.getDate() + n);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  };
  await page.getByRole('button', { name: '+ New countdown' }).click();
  const dialog = page.getByRole('dialog', { name: 'New countdown' });
  await dialog.getByLabel('Title').fill('Exam');
  await dialog.getByLabel('Date').fill(inDays(3));
  await dialog.getByLabel('Emoji').fill('📚');
  await dialog.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(dialog).toHaveCount(0);
  const item = page.getByRole('listitem').filter({ hasText: 'Exam' });
  await expect(item).toContainText('3 days');
  await expect(item).toHaveClass(/soon/);

  await item.getByRole('button', { name: 'Edit Exam' }).click();
  const edit = page.getByRole('dialog', { name: 'Edit countdown' });
  await edit.getByLabel('Title').fill('Holiday');
  await edit.getByLabel('Date').fill(inDays(30));
  await edit.getByRole('button', { name: 'Save', exact: true }).click();
  const holiday = page.getByRole('listitem').filter({ hasText: 'Holiday' });
  await expect(holiday).toContainText('30 days');
  await expect(holiday).not.toHaveClass(/soon/);

  page.once('dialog', (d) => d.accept());
  await holiday.getByRole('button', { name: 'Delete Holiday' }).click();
  await expect(page.getByText('No countdowns yet.')).toBeVisible();
});

test('the focus timer runs on timestamps and counts finished sessions (#58)', async ({ page }) => {
  await page.clock.install();
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Deep work');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('pomodoro');
  await page.getByRole('button', { name: /^Focus timer/ }).click();
  await page.getByLabel('Focus (min)').fill('1');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  const timer = page.getByRole('timer');
  await expect(timer).toHaveText('1:00');
  await expect(page.getByText('0 sessions in the last 7 days')).toBeVisible();
  await page.getByRole('button', { name: 'Start', exact: true }).click();
  await page.clock.runFor(30_000);
  await expect(timer).toHaveText('0:30');

  // A sleeping tab: the phase ends once, the break waits to be started.
  await page.clock.runFor(10 * 60_000);
  await expect(page.getByText('Short break', { exact: true })).toBeVisible();
  await expect(timer).toHaveText('5:00');
  await expect(page.getByText('1 sessions in the last 7 days')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Start', exact: true })).toBeVisible();
});

test('header with greeting, search pill, launcher and account menu (#189)', async ({ page }) => {
  // Any uncaught error (e.g. an effect loop) fails the test.
  const errors: string[] = [];
  let step = 'login';
  page.on('pageerror', (e) => errors.push(`${step}: ${e.message}`));
  page.on(
    'console',
    (m) => m.type() === 'error' && console.log(`[${step}] console: ${m.text().slice(0, 2000)}`),
  );
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  await expect(page.locator('time')).toBeVisible();

  // Height budget (#189, decided against the mockup): top bar, greeting,
  // search pill and tab pills before the first widget.
  const firstWidget = await page.locator('[data-widget-id]').first().boundingBox();
  expect(firstWidget?.y ?? 999).toBeLessThan(260);

  // Typing into the search pill continues in the quick search.
  step = 'search';
  await page.getByRole('searchbox', { name: 'Quick search (Ctrl+K or /)' }).pressSequentially('Dock');
  await expect(page.getByRole('combobox', { name: /Search bookmarks/ })).toHaveValue('Dock');
  await page.keyboard.press('Escape');

  // Launcher: filter across services and bookmarks; Esc returns the focus.
  step = 'launcher';
  const trigger = page.getByRole('button', { name: 'Services and bookmarks' });
  await trigger.click();
  const launcher = page.getByRole('dialog', { name: 'Services and bookmarks' });
  await expect(launcher.getByRole('heading', { name: 'Services' })).toBeVisible();
  await launcher.getByRole('searchbox', { name: /Filter services/ }).fill('docker');
  await expect(launcher.getByRole('link', { name: /Docker documentation/ })).toBeVisible();
  await expect(launcher.getByText('DuckDuckGo')).toHaveCount(0);
  await page.keyboard.press('Escape');
  await expect(launcher).toHaveCount(0);
  await expect(trigger).toBeFocused();

  // Account menu: keyboard operable, Esc returns the focus.
  step = 'account';
  const accountButton = page.getByRole('button', { name: 'Account menu', exact: true });
  await accountButton.click();
  await expect(page.getByRole('menuitem', { name: 'Edit', exact: true })).toBeFocused();
  await page.keyboard.press('ArrowDown');
  await expect(page.getByRole('menuitem', { name: 'Settings', exact: true })).toBeFocused();
  await page.keyboard.press('Escape');
  await expect(accountButton).toBeFocused();

  // The greeting addresses the user by name; the launcher shows only bookmarks.
  step = 'settings';
  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Appearance', exact: true }).click();
  await page.getByLabel('Your name in the greeting (optional)').fill('Tristan');
  await page.getByLabel(/The launcher/).selectOption('bookmarks');
  await page.getByRole('button', { name: 'Save appearance' }).click();
  await page.getByRole('button', { name: 'Close' }).click();
  await expect(page.locator('.hero h1')).toHaveText(/^Good (morning|afternoon|evening|night), Tristan$/);
  step = 'launcher bookmarks';
  await trigger.click();
  await expect(launcher.getByRole('heading', { name: 'Services' })).toHaveCount(0);
  await expect(launcher.getByRole('heading', { name: 'Bookmarks' })).toBeVisible();
  await page.keyboard.press('Escape');
  expect(errors).toEqual([]);
});

test('the smart search answers questions about the server (#156)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: /Search bookmarks/ });
  const answers = page.getByRole('region', { name: 'Answers' });

  await box.fill('cpu');
  await expect(answers.getByRole('article', { name: 'CPU' })).toContainText('%');

  await box.fill('wie voll ist die platte');
  await expect(answers.getByRole('article', { name: 'Disks' })).toContainText('free');

  // Service names with a typo; the sample data has DuckDuckGo.
  await box.fill('duckduckgoo');
  await expect(answers.getByRole('article', { name: 'DuckDuckGo' })).toBeVisible();

  // The weather module is disabled in this test: no weather answer.
  await box.fill('wetter');
  await expect(answers).toHaveCount(0);

  // "Open widget" jumps to the system widget on the Server tab.
  await box.fill('/ram');
  await answers
    .getByRole('article', { name: 'Memory' })
    .getByRole('button', { name: /Open widget/ })
    .click();
  await expect(page.getByRole('button', { name: 'Server', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  );
});

test('two-factor authentication can be enabled and is required at login', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Account', exact: true }).click();
  await page.getByRole('button', { name: 'Set up two-factor authentication' }).click();
  await expect(page.getByAltText('QR code for the authenticator app')).toBeVisible();
  const secret = (await page.getByRole('dialog').locator('code').textContent()) ?? '';
  await page.getByLabel('Code from the app').fill(totp(secret));
  await page.getByRole('button', { name: 'Turn on' }).click();
  await expect(page.getByText('Two-factor authentication is on.')).toBeVisible();
  await page.getByRole('button', { name: 'I saved them' }).click();
  await page.getByRole('button', { name: 'Close' }).click();

  await account(page, 'Log out');
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
