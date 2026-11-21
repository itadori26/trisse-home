import { expect, test } from '@playwright/test';
import { password, setupToken } from './support';

// The first start with the default SAMPLE_DATA=true, on a server of its own
// (see playwright.config.ts): the essential settings (#671), then the
// template choice (#670), then the tour. A template replaces the example
// tabs with a zoned tab.
test.use({ geolocation: { latitude: 53.14, longitude: 8.21 }, permissions: ['geolocation'] });

test('first start asks the essentials, offers the templates and Family replaces the example dashboard', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('Setup token').fill(setupToken);
  await page.getByLabel('Password (at least 12 characters)').fill(password);
  await page.getByLabel('Repeat password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  // The essentials: one question per screen, with the progress.
  const name = page.getByRole('dialog', { name: 'What should Waylight call you?' });
  await expect(name.getByText('Step 1 of 4')).toBeVisible();
  await name.getByLabel('Your name').fill('Sam');
  await name.getByRole('button', { name: 'Next' }).click();

  const place = page.getByRole('dialog', { name: 'Where are you?' });
  await expect(place.getByText('Step 2 of 4')).toBeVisible();
  await place.getByRole('button', { name: 'Use my location' }).click();
  await expect(place.getByRole('status')).toHaveText('Chosen: My location');
  await place.getByRole('button', { name: 'Next' }).click();

  const language = page.getByRole('dialog', { name: 'Language and time zone' });
  await expect(language.getByText(/time zone of this device/)).toBeVisible();
  await language.getByLabel('Language').selectOption('en');
  await language.getByRole('button', { name: 'Next' }).click();

  const look = page.getByRole('dialog', { name: 'How should Waylight look?' });
  await expect(look.getByText('Step 4 of 4')).toBeVisible();
  await look.getByRole('radio', { name: 'Dark' }).check();
  // The page previews the choice before it is saved.
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  await look.getByRole('button', { name: 'Next' }).click();

  const picker = page.getByRole('dialog', { name: 'What do you use Waylight for?' });
  await expect(picker).toBeVisible();
  await expect(picker.getByText('it replaces the example dashboard')).toBeVisible();
  // The tour waits for the choice.
  await expect(page.getByRole('dialog', { name: 'Welcome to Waylight 👋' })).toHaveCount(0);

  await picker.getByText('Family', { exact: true }).click();
  // The place from the essentials is used, so the location is not asked again.
  await picker.getByRole('button', { name: 'Use this template' }).click();

  const tour = page.getByRole('dialog', { name: 'Welcome to Waylight 👋' });
  await expect(tour).toBeVisible();
  await tour.getByRole('button', { name: 'Skip tour' }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(page.locator('.hero h1')).toHaveText(/, Sam$/);

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

  // The answers are stored where the settings pages store them.
  const stored = (await (await page.request.get('/api/appearance')).json()) as Record<string, unknown>;
  expect(stored).toMatchObject({ greetingName: 'Sam', language: 'en', theme: 'dark' });
  const places = (await (await page.request.get('/api/places')).json()) as { name: string; lat: number }[];
  expect(places).toEqual([expect.objectContaining({ name: 'Home', lat: 53.14 })]);
  // The example data of the modules stays.
  expect(((await (await page.request.get('/api/todos')).json()) as unknown[]).length).toBeGreaterThan(0);
});
