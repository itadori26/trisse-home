import { defineConfig, devices } from '@playwright/test';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

/** Port of the throw-away server started for the smoke test. */
const port = Number(process.env.E2E_PORT ?? 18091);

/** Fixed setup token so the test can complete the first-run setup. */
export const setupToken = 'e2e-setup-token-0123456789';

const dataDir = mkdtempSync(join(tmpdir(), 'trisse-home-e2e-'));

export default defineConfig({
  testDir: 'e2e',
  timeout: 30_000,
  retries: 0,
  workers: 1,
  fullyParallel: false,
  reporter: 'list',
  use: { baseURL: `http://127.0.0.1:${port}`, trace: 'retain-on-failure' },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: '../bin/trisse-home',
    url: `http://127.0.0.1:${port}/healthz`,
    reuseExistingServer: false,
    env: {
      LISTEN_ADDR: `127.0.0.1:${port}`,
      DATABASE_PATH: join(dataDir, 'e2e.db'),
      SECRET_KEY: 'MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=',
      SETUP_TOKEN: setupToken,
      ALLOWED_CIDRS: '127.0.0.0/8',
      LOG_LEVEL: 'warn',
      MODULES_DISABLED: 'docker,adguard,calendar,rss,weather',
    },
  },
});
