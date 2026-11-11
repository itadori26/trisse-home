import { defineConfig, devices } from '@playwright/test';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

/**
 * Generates the README screenshots from a throw-away DEMO_MODE instance
 * (sample data, fixture integrations, no internet access needed):
 *   make build && cd web && npm run screenshots
 */
const port = Number(process.env.SCREENSHOT_PORT ?? 18092);
/** Set to screenshot an already running demo (e.g. the release image). */
const external = process.env.SCREENSHOT_BASE_URL;
const dataDir = mkdtempSync(join(tmpdir(), 'trisse-home-shots-'));

export default defineConfig({
  testDir: 'e2e/screenshots',
  timeout: 90_000,
  workers: 1,
  reporter: 'list',
  use: { baseURL: external ?? `http://127.0.0.1:${port}` },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: external
    ? undefined
    : {
        command: '../bin/trisse-home',
        url: `http://127.0.0.1:${port}/healthz`,
        reuseExistingServer: false,
        env: {
          LISTEN_ADDR: `127.0.0.1:${port}`,
          DATABASE_PATH: join(dataDir, 'demo.db'),
          SECRET_KEY: 'MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=',
          ALLOWED_CIDRS: '127.0.0.0/8',
          LOG_LEVEL: 'warn',
          DEMO_MODE: 'true',
          DOCKER_PROXY_URL: 'http://docker-socket-proxy:2375',
        },
      },
});
