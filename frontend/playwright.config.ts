import { defineConfig, devices } from '@playwright/test'

const isDEngagementRun = process.argv.some((argument) => argument.includes('d-engagement'))
const isMemberCRun = process.argv.some((argument) => argument.includes('member-c-content'))
if (isDEngagementRun) process.env.E2E_SKIP_UC13_SETUP = '1'
if (isMemberCRun) process.env.E2E_MEMBER_C_RUN = '1'
const backendConfig = process.env.E2E_BACKEND_CONFIG ?? 'etc/config.yaml'

export default defineConfig({
  testDir: './e2e',
  outputDir: isMemberCRun ? '/tmp/danmakustream-member-c-test-results' : 'test-results',
  globalSetup: './e2e/global-setup.ts',
  globalTeardown: './e2e/global-teardown.ts',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [
    ['list'],
    ['html', {
      outputFolder: isDEngagementRun
        ? '../docs/testing/reports/engagement-e2e'
        : isMemberCRun
          ? '../docs/testing/reports/UC02-03-04-12-e2e-report'
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
      command: isMemberCRun
        ? 'npm run build && npm run preview -- --host 127.0.0.1 --port 5173'
        : 'npm run dev -- --host 127.0.0.1',
      url: 'http://127.0.0.1:5173',
      reuseExistingServer: true,
      timeout: 60_000,
    },
  ],
})
