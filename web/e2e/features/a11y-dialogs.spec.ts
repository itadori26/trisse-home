import { expect, test } from '@playwright/test';
import { expectAccessible } from '../a11y';
import { account, password } from '../support';

// Automated WCAG 2.2 AA checks (#113) for the dialogs that the main-view
// test in smoke.spec.ts does not open: the tab dialogs, a widget's settings,
// the background picker, and the dashboard on a phone. Works on a tab of its
// own and removes it at the end, so the other specs see the same state.
test('tab, widget and background dialogs pass automated WCAG 2.2 AA checks (#113)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await expectAccessible(page, 'new tab dialog', 'dialog[open]');
  await page.getByRole('textbox', { name: 'Name' }).fill('A11y');
  await page.getByRole('button', { name: 'Create' }).click();

  // A new tab offers rule presets once (#673); the choice is a plain rule.
  const offer = page.getByTestId('tab-rule-offer');
  await expect(offer).toContainText('Open “A11y” automatically?');
  await expectAccessible(page, 'tab rule offer', '[data-testid="tab-rule-offer"]');
  await offer.getByRole('button', { name: 'Weekdays, 8–17' }).click();
  await expect(offer.getByRole('status')).toContainText('now opens automatically');
  await offer.getByRole('button', { name: 'Got it' }).click();
  await expect(offer).toHaveCount(0);

  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('button', { name: /^Clock/ }).click();
  await page.getByRole('button', { name: 'Save', exact: true }).click();

  await page.getByRole('button', { name: 'Widget settings' }).first().click();
  await expectAccessible(page, 'widget settings', 'dialog[open]');
  await page.keyboard.press('Escape');

  const tabs = page.getByRole('navigation', { name: 'Dashboard tabs' });
  const tabMenu = () =>
    tabs.locator('.tab', { hasText: 'A11y' }).getByRole('button', { name: 'Tab actions' }).click();
  await tabMenu();
  await expectAccessible(page, 'tab menu', '[role="menu"]');
  await page.getByRole('menuitem', { name: 'Open by default' }).click();
  await expect(page.getByRole('dialog').getByLabel('from')).toHaveValue('08:00');
  await expectAccessible(page, 'tab rules dialog', 'dialog[open]');
  await page.keyboard.press('Escape');

  // The background picker (#277), with its scope switch and swatches.
  await page
    .getByRole('toolbar', { name: 'Edit actions' })
    .getByRole('button', { name: 'Background' })
    .click();
  const picker = page.getByRole('dialog', { name: 'Background' });
  await expect(picker.getByRole('radio', { name: 'All tabs' })).toBeVisible();
  await page.waitForLoadState('networkidle');
  await expectAccessible(page, 'background picker', 'dialog[open]');
  await page.keyboard.press('Escape');

  await tabMenu();
  await page.getByRole('menuitem', { name: 'Delete tab' }).click();
  await page.getByRole('dialog', { name: 'Delete tab' }).getByRole('button', { name: 'Delete tab' }).click();
  await expect(tabs.getByRole('button', { name: 'A11y', exact: true })).toBeHidden();
  await account(page, 'Done');

  // Phones get a single column with their own header (#92, #226).
  await page.setViewportSize({ width: 390, height: 844 });
  await page.reload();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.waitForLoadState('networkidle');
  await expectAccessible(page, 'dashboard on a phone');
});
