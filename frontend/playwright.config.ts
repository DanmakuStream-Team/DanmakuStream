import { defineConfig, devices } from '@playwright/test'

const isDEngagementRun = process.argv.some((argument) => argument.includes('d-engagement'))
const isMemberCRun = process.argv.some((argument) => argument.includes('member-c-content'))
const isUC01UC06Run = process.argv.some((argument) => argument.includes('uc01-user') || argument.includes('uc06-library'))
if (isDEngagementRun) process.env.E2E_SKIP_UC13_SETUP = '1'
if (isMemberCRun) process.env.E2E_MEMBER_C_RUN = '1'
if (isUC01UC06Run) process.env.E2E_SKIP_UC13_SETUP = '1'
const backendConfig = process.env.E2E_BACKEND_CONFIG ?? 'etc/config.yaml'
const useGateway = process.env.E2E_USE_GATEWAY === '1'

// ── E2E 网关链路约定 ────────────────────────────────────────────────────
// 默认（本地 go run + vite dev 模式）：Vite proxy 直连本机 backend(8080)，
//                                      与单体后端保持一致、无需额外起 compose。
// 设 E2E_USE_GATEWAY=1 时：              Vite proxy → 宿主 nginx-gateway(8888)
//                                      → backend:8080（容器内），
// 此时需要先 `docker compose up -d` 起 mysql/srs/backend/nginx-gateway。
const gatewayTarget = useGateway ? 'http://127.0.0.1:8888' : 'http://127.0.0.1:8080'
const gatewayWSTarget = useGateway ? 'ws://127.0.0.1:8888' : 'ws://127.0.0.1:8080'
process.env.VITE_DEV_GATEWAY_TARGET = gatewayTarget
process.env.VITE_DEV_GATEWAY_WS_TARGET = gatewayWSTarget

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
        : isUC01UC06Run
          ? '../docs/testing/reports/UC01-UC06-e2e-report'
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
    navigationTimeout: 30_000,
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
      env: {
        VITE_DEV_GATEWAY_TARGET: gatewayTarget,
        VITE_DEV_GATEWAY_WS_TARGET: gatewayWSTarget,
      },
    },
  ],
})
