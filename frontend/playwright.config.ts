import { defineConfig, devices } from '@playwright/test';

// The end-to-end checks of the screen (ADR-0017): the real API and the Vite
// dev server, in Chromium only. `make e2e` builds the API first.
//
// The ports are their own, so that a dev loop on 8080 and 5173 does not answer
// in their place; neither server is reused.
const API_PORT = 18080;
const WEB_PORT = 15173;
const API = `http://127.0.0.1:${API_PORT}`;
const WEB = `http://127.0.0.1:${WEB_PORT}`;

// PW_CHROMIUM_PATH points at a Chromium already on the machine, for when the
// one this version of Playwright downloads is not there (docs/dev/harness.md).
const chromium = process.env.PW_CHROMIUM_PATH;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: WEB,
    trace: 'off',
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        ...(chromium ? { launchOptions: { executablePath: chromium } } : {}),
      },
    },
  ],
  webServer: [
    {
      command: './bin/server',
      cwd: '../backend',
      env: { PORT: String(API_PORT) },
      url: `${API}/api/v1/health`,
      reuseExistingServer: false,
    },
    {
      command: `npx vite --host 127.0.0.1 --port ${WEB_PORT} --strictPort`,
      env: { API_UPSTREAM: API },
      url: WEB,
      reuseExistingServer: false,
    },
  ],
});
