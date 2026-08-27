import { defineConfig, devices } from '@playwright/test'

const isDEngagementRun = process.argv.some((argument) => argument.includes('d-engagement'))
const isUserDomainRun = process.argv.some((argument) => argument.includes('user-domain'))
if (isDEngagementRun || isUserDomainRun) process.env.E2E_SKIP_UC13_SETUP = '1'
if (isUserDomainRun) process.env.E2E_RUN_USER_DOMAIN = '1'
const backendConfig = process.env.E2E_BACKEND_CONFIG ?? 'etc/config.yaml'

export default defineConfig({
  testDir: './e2e',
  globalSetup: './e2e/global-setup.ts',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [
    ['list'],
    ['html', {
      outputFolder: isDEngagementRun
        ? '../docs/testing/reports/engagement-e2e'
        : isUserDomainRun
          ? '../docs/testing/reports/user-domain-e2e'
        : '../docs/testing/reports/uc13-e2e',
      open: 'never',
    }],
  ],
  use: {
    baseURL: 'http://127.0.0.1:5173',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10_000,
    navigationTimeout: 15_000,
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: `go run api/main.go -f "${backendConfig}"`,
      cwd: '../backend',
      url: 'http://127.0.0.1:8080/api/v1/live',
      reuseExistingServer: true,
      timeout: 60_000,
    },
    {
      command: 'npm run dev -- --host 127.0.0.1',
      url: 'http://127.0.0.1:5173',
      reuseExistingServer: true,
      timeout: 60_000,
    },
  ],
})
