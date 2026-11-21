import { expect, test } from '@playwright/test';
import { expectAccessible } from './a11y';
import { account, password, openSettings } from './support';

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
  // The new place is announced through the app's shared live region (#257).
  const live = page.locator('[aria-live="polite"][aria-atomic="true"]');
  await page.keyboard.press('ArrowRight');
  await expect(live).toHaveText(/^Clock moved: column 2, row 1/);
  await page.keyboard.press('Shift+ArrowRight');
  await expect(live).toHaveText(/^Clock resized: column 2, row 1, \d+ × \d+ cells/);
  await expect(page.getByRole('group', { name: /^Clock, column 2, row 1/ })).toBeFocused();
  await account(page, 'Done');

  await page.reload();
  await page.getByRole('button', { name: 'Study' }).click();
  await account(page, 'Edit');
  await expect(page.getByRole('group', { name: /^Clock, column 2, row 1/ })).toBeVisible();
});

test('quick search finds bookmarks with the keyboard', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  // Ctrl+K focuses the hero field; the results open below it (#222).
  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  const results = page.getByRole('listbox', { name: 'Results' });
  await expect(box).toBeFocused();
  await box.fill('docker docs');
  await expect(page.getByRole('option', { name: /Docker documentation/ })).toHaveAttribute(
    'aria-selected',
    'true',
  );
  await expect(box).toHaveAttribute('aria-expanded', 'true');
  // Without a match the web search is the only row.
  await box.fill('nothing-matches-this');
  await expect(page.getByRole('option', { name: /nothing-matches-this/ })).toHaveAttribute(
    'aria-selected',
    'true',
  );
  // The first Esc closes the panel and keeps the text, the second clears the field.
  await page.keyboard.press('Escape');
  await expect(results).toBeHidden();
  await expect(box).toHaveValue('nothing-matches-this');
  await page.keyboard.press('Escape');
  await expect(box).toHaveValue('');
  await expect(box).not.toBeFocused();

  await page.keyboard.press('/');
  await expect(box).toBeFocused();
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
    const cache = await caches.open('waylight-data-v1');
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
  await openSettings(page, 'Connections & services', 'Proxmox');
  await expect(page.getByLabel('Proxmox VE URL')).toBeVisible();
});

test('settings: start page with areas, keyboard navigation and phones (#600)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Settings');
  const dialog = page.getByRole('dialog', { name: 'Settings' });
  const nav = dialog.getByRole('navigation', { name: 'Settings areas' });
  // The start page lists the areas as tiles.
  const tiles = dialog.getByRole('list', { name: 'Settings areas' });
  await expect(tiles.getByRole('button', { name: /^Connections & services/ })).toBeVisible();
  // The backup traffic light (#677): the job backs up right after the start.
  await expect(tiles.getByRole('button', { name: /^Data & backup/ })).toContainText(
    /Last backup today .*, all good/,
  );
  await tiles.getByRole('button', { name: /^Connections & services/ }).click();
  await expect(dialog.getByRole('region', { name: 'Connections & services', exact: true })).toBeVisible();
  // An area lists its sections; the open section is marked in the navigation.
  await dialog.getByRole('region').getByRole('button', { name: 'Proxmox', exact: true }).click();
  await expect(page.getByLabel('Proxmox VE URL')).toBeVisible();
  await expect(nav.getByRole('button', { name: 'Proxmox', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  );
  // Arrow keys move between entries, Enter selects.
  await nav.getByRole('button', { name: 'Overview', exact: true }).click();
  await nav.getByRole('button', { name: 'Overview', exact: true }).focus();
  await page.keyboard.press('ArrowDown');
  await expect(nav.getByRole('button', { name: 'My account', exact: true })).toBeFocused();
  await page.keyboard.press('Enter');
  await expect(dialog.getByRole('region', { name: 'Account', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Close' }).click();

  // Phones: one page at a time with a way back instead of the navigation.
  await page.setViewportSize({ width: 390, height: 844 });
  await account(page, 'Settings');
  await expect(nav).toBeHidden();
  await tiles.getByRole('button', { name: /^Dashboard & widgets/ }).click();
  await dialog.getByRole('region').getByRole('button', { name: 'Search', exact: true }).click();
  await expect(dialog.getByRole('region', { name: 'Search', exact: true })).toBeVisible();
  await dialog.getByRole('button', { name: '← Dashboard & widgets' }).click();
  await dialog.getByRole('button', { name: '← Overview' }).click();
  await expect(tiles).toBeVisible();
});

test('a guest sees only the shared tab and cannot edit', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await openSettings(page, 'People & access', 'Guests');
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
  await expect(page.getByRole('button', { name: /^My account/ }).first()).toBeVisible();
  await expect(page.getByRole('button', { name: /^People & access/ })).toHaveCount(0);
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
  await expect(page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' })).toBeFocused();
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

test('the search field captures items and its options live in the settings (#105, #67, #221)', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  // The search field next to the greeting is the search (#221, #222).
  const field = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });

  await field.fill('t Milch kaufen morgen');
  await expect(
    page.getByRole('option', { name: /create to-do “Milch kaufen” \(Tomorrow\)/ }),
  ).toHaveAttribute('aria-selected', 'true');
  await field.press('Enter');
  await expect(page.getByText('To-do “Milch kaufen” created')).toBeVisible();
  const todos = await (await page.request.get('/api/todos')).json();
  expect(JSON.stringify(todos)).toContain('Milch kaufen');

  // Ctrl+Enter adds the same text as the other kind (mockup "E", #222).
  await field.fill('t Einkaufszettel');
  await expect(page.getByRole('option', { name: /create note “Einkaufszettel”/ })).toBeVisible();
  await field.press('Control+Enter');
  await expect(page.getByText('Note “Einkaufszettel” created')).toBeVisible();

  await field.fill('n Idee Fahrradtour');
  await field.press('Enter');
  await expect(page.getByText('Note “Idee Fahrradtour” created')).toBeVisible();
  await page.getByRole('status').filter({ hasText: 'created' }).getByRole('button', { name: 'Undo' }).click();

  await field.fill('b example.org Beispiel');
  await field.press('Enter');
  await expect(page.getByText('Bookmark “Beispiel” saved')).toBeVisible();
  await expect(page.getByRole('status').getByRole('link', { name: 'Open' })).toHaveAttribute(
    'href',
    'https://example.org/',
  );

  // The palette's own prefixes still work.
  await field.fill('todo: Steuer heute');
  await expect(page.getByRole('option', { name: /create to-do “Steuer” \(Today\)/ })).toBeVisible();
  await field.press('Enter');
  await expect(page.getByText('To-do “Steuer” created')).toBeVisible();
  await page.keyboard.press('Escape');

  // The prefixes are a search setting now, not a widget option.
  await openSettings(page, 'Dashboard & widgets', 'Search');
  const todoPrefix = page.getByLabel('Prefix for to-dos');
  await expect(todoPrefix).toHaveValue('t');
  await todoPrefix.fill('a');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByText('Saved', { exact: true })).toBeVisible();
  await page.keyboard.press('Escape');
  await field.fill('a Kuchen backen');
  await expect(page.getByRole('option', { name: /create to-do “Kuchen backen”/ })).toBeVisible();
  await page.keyboard.press('Escape');
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

  // The text stays in the hero field while the results open below it (#222).
  step = 'search';
  const field = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  await field.click();
  await field.pressSequentially('Dock');
  await expect(field).toHaveValue('Dock');
  await expect(field).toBeFocused();
  await expect(page.getByRole('listbox', { name: 'Results' })).toBeVisible();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await page.keyboard.press('Escape');
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
  await openSettings(page, 'Appearance & language', 'Appearance');
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

test('widgets poll at their interval, not in a loop (#219)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  // Every widget loads once and then every 30 s or more; a polling loop shows
  // up as the same endpoint dozens of times per second.
  const counts = new Map<string, number>();
  page.on('request', (r) => {
    const path = new URL(r.url()).pathname;
    if (path.startsWith('/api/')) counts.set(path, (counts.get(path) ?? 0) + 1);
  });
  await page.getByRole('button', { name: 'Server', exact: true }).click();
  await page.waitForTimeout(6000);
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  await page.waitForTimeout(4000);
  const busiest = [...counts].sort((a, b) => b[1] - a[1]);
  expect(busiest.length, 'the dashboard made no API requests at all').toBeGreaterThan(0);
  expect(busiest[0]?.[1] ?? 0, JSON.stringify(busiest.slice(0, 5))).toBeLessThanOrEqual(4);
});

test('a kiosk link shows the chosen tab read-only without controls (#50)', async ({ page, browser }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await openSettings(page, 'People & access', 'Kiosk displays');
  await page.getByLabel('Name (e.g. flur-tablet)').fill('flur');
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('checkbox', { name: 'Home' }).check();
  await dialog.getByLabel('Tab rotation').selectOption('30');
  await page.getByRole('button', { name: 'Create display' }).click();
  await expect(page.getByRole('img', { name: 'QR code of the kiosk link' })).toBeVisible();
  const link = (await dialog.locator('code').textContent()) ?? '';
  expect(link).toMatch(/\/kiosk#.+/);
  await expect(dialog.getByRole('listitem').getByText('next tab every 30 s')).toBeVisible();

  const context = await browser.newContext();
  const display = await context.newPage();
  await display.goto(link);
  await expect(display.getByTestId('dashboard')).toBeVisible();
  await expect(display).toHaveURL(/\/$/); // the token left the address bar
  await expect(display.getByRole('button', { name: 'Account menu' })).toBeHidden();
  await expect(display.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' })).toBeHidden();
  expect((await display.request.get('/api/auth/sessions')).status()).toBe(403);
  await context.close();
  await dialog.getByRole('button', { name: 'Done' }).click();
  page.once('dialog', (d) => d.accept());
  await dialog.getByRole('button', { name: 'Delete' }).first().click();
  await expect(dialog.getByText('No kiosk displays yet.')).toBeVisible();
  await page.getByRole('button', { name: 'Close' }).click();
});

test('the header never scrolls the page sideways on a phone, even with many tabs (#226)', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const status = await (await page.request.get('/api/auth/status')).json();
  const headers = { 'X-CSRF-Token': status.csrfToken as string };
  for (const name of [
    'Einkaufen',
    'Familie',
    'Server-Raum',
    'Studium',
    'Garten',
    'Urlaubsplanung',
    'Fotos',
  ]) {
    await page.request.post('/api/dashboard/tabs', { data: { name }, headers });
  }
  await page.setViewportSize({ width: 390, height: 844 });
  await page.reload();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
  // The tab bar scrolls inside itself; the last tab can be reached.
  await expect(page.getByRole('button', { name: 'Fotos', exact: true })).toBeAttached();
});

test('the main views pass automated WCAG 2.2 AA checks (#113)', async ({ page, browser }) => {
  const context = await browser.newContext();
  const anonymous = await context.newPage();
  await anonymous.goto('/');
  await expect(anonymous.getByLabel('Password')).toBeVisible();
  await expectAccessible(anonymous, 'login');
  await context.close();

  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.waitForLoadState('networkidle');
  await expectAccessible(page, 'dashboard');

  await account(page, 'Edit');
  await expectAccessible(page, 'edit mode');
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await expectAccessible(page, 'widget picker', 'dialog[open]');
  await page.keyboard.press('Escape');
  await account(page, 'Done');

  await account(page, 'Settings');
  const settings = page.getByRole('dialog');
  await expectAccessible(page, 'settings', 'dialog[open]');
  // Every area and section, including the ones of self-registering features.
  const nav = settings.getByRole('navigation', { name: 'Settings areas' });
  const areas = (await nav.locator('button[data-nav]:not(.sub)').allTextContents()).map((a) => a.trim());
  for (const area of areas) {
    await nav.getByRole('button', { name: area, exact: true }).click();
    await page.waitForLoadState('networkidle');
    await expectAccessible(page, `settings ${area}`, 'dialog[open]');
    const sections = (await nav.locator('button.sub').allTextContents()).map((s) => s.trim());
    for (const section of sections) {
      await nav.getByRole('button', { name: section, exact: true }).click();
      await page.waitForLoadState('networkidle');
      await expectAccessible(page, `settings ${area} › ${section}`, 'dialog[open]');
    }
  }
  await page.getByRole('button', { name: 'Close' }).click();

  await page.keyboard.press('?');
  await expectAccessible(page, 'shortcuts', 'dialog[open]');
  await page.keyboard.press('Escape');

  await page.emulateMedia({ colorScheme: 'dark' });
  await expectAccessible(page, 'dashboard dark');
});

test('?focus=search puts the cursor into the search field (browser extension, #65)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.goto('/?focus=search');
  await expect(page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' })).toBeFocused();
  // The parameter is dropped, so a reload does not grab the focus again.
  await expect(page).toHaveURL(/\/$/);
});

test('an empty tab offers "Add widget"; a click on a title adds no undo step (#276, #272)', async ({
  page,
}) => {
  const loops: string[] = [];
  page.on('console', (m) => m.text().includes('effect_update_depth_exceeded') && loops.push(m.text()));
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Leer');
  await page.getByRole('button', { name: 'Create' }).click();
  await account(page, 'Done');
  await expect(page.locator('p.empty')).toContainText('This tab is empty.');
  // The button starts edit mode and opens the picker.
  await page.locator('p.empty').getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('clock');
  await page.getByRole('button', { name: /^Clock/ }).click();
  // The clock has settings: its dialog opens after adding.
  await page.getByRole('dialog', { name: 'Clock settings' }).getByRole('button', { name: 'Close' }).click();
  const undo = page.getByRole('button', { name: '↶ Undo' });
  await expect(undo).toBeEnabled();
  // A click on the title bar without dragging must not add a step (#272):
  // one undo removes the clock again.
  const clock = page.getByRole('group', { name: /^Clock, column 1, row 1/ });
  const box = await clock.boundingBox();
  if (!box) throw new Error('no clock');
  await page.mouse.click(box.x + box.width / 2, box.y + 12);
  await undo.click();
  await expect(clock).toBeHidden();
  await expect(undo).toBeDisabled();
  await account(page, 'Done');
  expect(loops).toEqual([]);
});

test('edit mode: Done button, Esc, tab menu, rename by double-click, delete in a dialog (#271, #274)', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await expect(page.getByText('Edit mode: drag a title bar to move')).toBeVisible();
  const bar = page.getByRole('toolbar', { name: 'Edit actions' });
  await bar.getByRole('button', { name: '✓ Done' }).click();
  await expect(bar).toBeHidden();

  await account(page, 'Edit');
  await page.keyboard.press('Escape');
  await expect(bar).toBeHidden();

  // Tab actions live in one menu; double-click renames.
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Menü');
  await page.getByRole('button', { name: 'Create' }).click();
  const tabs = page.getByRole('navigation', { name: 'Dashboard tabs' });
  await tabs.getByRole('button', { name: 'Menü', exact: true }).dblclick();
  await page.getByRole('textbox', { name: 'Name' }).fill('Umbenannt');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(tabs.getByRole('button', { name: 'Umbenannt', exact: true })).toBeVisible();
  await tabs.locator('.tab', { hasText: 'Umbenannt' }).getByRole('button', { name: 'Tab actions' }).click();
  const menu = page.getByRole('menu', { name: 'Actions for tab Umbenannt' });
  await expect(menu.getByRole('menuitem', { name: 'Rename tab' })).toBeFocused();
  await page.keyboard.press('ArrowDown');
  await page.keyboard.press('Escape'); // closes only the menu, edit mode stays
  await expect(menu).toBeHidden();
  await expect(bar).toBeVisible();
  await tabs.locator('.tab', { hasText: 'Umbenannt' }).getByRole('button', { name: 'Tab actions' }).click();
  await menu.getByRole('menuitem', { name: 'Delete tab' }).click();
  const confirm = page.getByRole('dialog', { name: 'Delete tab' });
  await expect(confirm).toContainText('Delete the tab “Umbenannt” and all its widgets?');
  await confirm.getByRole('button', { name: 'Delete tab' }).click();
  await expect(tabs.getByRole('button', { name: 'Umbenannt', exact: true })).toBeHidden();
  await bar.getByRole('button', { name: '✓ Done' }).click();
});
