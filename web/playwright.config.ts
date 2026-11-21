import { defineConfig, devices } from '@playwright/test';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

/** Port of the throw-away server started for the smoke test. */
const port = Number(process.env.E2E_PORT ?? 18091);

/** Fixed setup token so the test can complete the first-run setup. */
export const setupToken = 'e2e-setup-token-0123456789';

const dataDir = mkdtempSync(join(tmpdir(), 'waylight-e2e-'));

/**
 * Port of a second, untouched server for the first start (#670): the shared
 * account of the chain has made its template choice already.
 */
const firstStartPort = port + 1;

/** Environment both servers share. */
const serverEnv = {
  SECRET_KEY: 'MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=',
  SETUP_TOKEN: setupToken,
  ALLOWED_CIDRS: '127.0.0.0/8',
  LOG_LEVEL: 'warn',
  // calendar stays on: unconfigured it makes no outbound requests, and #163 tests its split view.
  MODULES_DISABLED: 'docker,adguard,rss,weather,markets,sports,electricity',
};

const chrome = { ...devices['Desktop Chrome'] };

export default defineConfig({
  testDir: 'e2e',
  testIgnore: 'screenshots/**',
  timeout: 90_000,
  retries: 0,
  workers: 1,
  fullyParallel: false,
  reporter: 'list',
  use: { baseURL: `http://127.0.0.1:${port}`, trace: 'retain-on-failure' },
  // The specs share one server and one account, so they run as a chain of
  // projects (#235): setup creates the account, the general flows and the
  // feature specs (one file each, e2e/features/) use it, and two-factor
  // runs last because it turns on 2FA for that account. A new feature test
  // is a new file in e2e/features/ and touches no shared line.
  projects: [
    { name: 'setup', testMatch: 'setup.spec.ts', use: chrome },
    { name: 'general', testMatch: 'smoke.spec.ts', dependencies: ['setup'], use: chrome },
    { name: 'features', testMatch: 'features/**/*.spec.ts', dependencies: ['general'], use: chrome },
    { name: 'twofactor', testMatch: 'twofactor.spec.ts', dependencies: ['features'], use: chrome },
    // The browser extension in local mode (#291) starts its own Chromium and needs no account.
    { name: 'extension', testMatch: 'extension.spec.ts' },
    // The first start with sample data (#670) on its own fresh server.
    {
      name: 'first-start',
      testMatch: 'first-start.spec.ts',
      use: { ...chrome, baseURL: `http://127.0.0.1:${firstStartPort}` },
    },
  ],
  webServer: [
    {
      command: '../bin/waylight',
      url: `http://127.0.0.1:${port}/healthz`,
      reuseExistingServer: false,
      env: { ...serverEnv, LISTEN_ADDR: `127.0.0.1:${port}`, DATABASE_PATH: join(dataDir, 'e2e.db') },
    },
    {
      command: '../bin/waylight',
      url: `http://127.0.0.1:${firstStartPort}/healthz`,
      reuseExistingServer: false,
      env: {
        ...serverEnv,
        LISTEN_ADDR: `127.0.0.1:${firstStartPort}`,
        DATABASE_PATH: join(dataDir, 'first-start.db'),
        SAMPLE_DATA: 'true',
      },
    },
  ],
});
