import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test.describe('places in widget settings', () => {
  // The service worker would answer the API itself; page.route only sees requests without it.
  test.use({
    serviceWorkers: 'block',
    permissions: ['geolocation'],
    geolocation: { latitude: 53.143549, longitude: 8.214649 },
  });

  test('use my location saves a rounded place and lists the stops near it (#693, #694)', async ({ page }) => {
    let asked = '';
    await page.route(
      (u) => u.pathname === '/api/transit/stops/near',
      (r) => {
        asked = new URL(r.request().url()).search;
        return r.fulfill({
          json: [{ id: 'de-DELFI_de:03403:16750', name: 'Oldenburg Hbf', lat: 53.144, lon: 8.222 }],
        });
      },
    );

    // Keeps the saved widget from asking the real journey planner.
    await page.route(
      (u) => u.pathname === '/api/transit/departures',
      (r) => r.fulfill({ json: [] }),
    );

    await page.goto('/');
    await page.getByLabel('Password').fill(password);
    await page.getByRole('button', { name: 'Log in' }).click();
    await expect(page.getByTestId('dashboard')).toBeVisible();

    await account(page, 'Edit');
    await page.getByRole('button', { name: '+ Tab' }).click();
    await page.getByRole('textbox', { name: 'Name' }).fill('Nearby');
    await page.getByRole('button', { name: 'Create' }).click();
    await page.getByRole('button', { name: '+ Add widget' }).click();
    await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('departures');
    await page.getByRole('button', { name: /^Departures/ }).click();

    const settings = page.getByRole('dialog', { name: 'Departures settings' });
    const locate = settings.getByRole('button', { name: 'Use my location' });
    await expect(locate).toHaveAccessibleDescription(/rounded to about 100 m/);
    await locate.click();

    // The place is saved and chosen; only its rounded coordinates leave the browser.
    await expect(settings.getByLabel('Saved place')).toHaveValue(/.+/);
    await settings.getByRole('button', { name: 'Oldenburg Hbf' }).click();
    expect(asked).toBe('?lat=53.144&lon=8.215');
    await expect(settings.getByText('Stop: Oldenburg Hbf')).toBeVisible();
    await settings.getByRole('button', { name: 'Save', exact: true }).click();
    await expect(settings).toHaveCount(0);
    await account(page, 'Done');
  });
});
