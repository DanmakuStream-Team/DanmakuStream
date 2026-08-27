import { defineConfig, devices } from '@playwright/test'

const isDEngagementRun = process.argv.some((argument) => argument.includes('d-engagement'))
if (isDEngagementRun) process.env.E2E_SKIP_UC13_SETUP = '1'
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
  globalSetup: './e2e/global-setup.ts',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [
    ['list'],
    ['html', {
      outputFolder: isDEngagementRun
        ? '../docs/testing/reports/engagement-e2e'
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
      env: {
        VITE_DEV_GATEWAY_TARGET: gatewayTarget,
        VITE_DEV_GATEWAY_WS_TARGET: gatewayWSTarget,
      },
    },
  ],
})
