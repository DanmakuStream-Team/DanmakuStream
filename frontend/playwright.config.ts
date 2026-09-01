import { defineConfig, devices } from '@playwright/test'

// 统一 E2E 配置：不再按成员/套件区分运行模式。
// - 全部 spec 一次跑完（`npm run test:e2e`；指定单个文件：`npm run test:e2e -- e2e/uc13-admin.spec.ts`）
// - global-setup 无条件准备三域测试数据（账号与数据按前缀隔离、可重复执行）
// - 未实现的骨架 spec 以 test.skip(true, ...) 显式跳过，实现后删除该行即自动纳入
//
// 网关链路约定：默认 Vite proxy 直连本机 backend(8080)；
// 设 E2E_USE_GATEWAY=1 时走宿主 nginx-gateway(8888)（需先 docker compose up -d）。

const MICRO = process.env.E2E_MICROSERVICES === '1'
const backendConfig = process.env.E2E_BACKEND_CONFIG ?? 'etc/config.yaml'
const useGateway = process.env.E2E_USE_GATEWAY === '1'
const gatewayTarget = useGateway ? 'http://127.0.0.1:8888' : 'http://127.0.0.1:8080'
const gatewayWSTarget = useGateway ? 'ws://127.0.0.1:8888' : 'ws://127.0.0.1:8080'
process.env.VITE_DEV_GATEWAY_TARGET = gatewayTarget
process.env.VITE_DEV_GATEWAY_WS_TARGET = gatewayWSTarget

export default defineConfig({
  testDir: './e2e',
  outputDir: 'test-results',
  globalSetup: './e2e/global-setup.ts',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [
    ['list'],
    ['html', { outputFolder: '../docs/testing/reports/e2e', open: 'never' }],
  ],
  use: {
    baseURL: MICRO ? (process.env.E2E_BASE_URL ?? 'http://127.0.0.1') : 'http://127.0.0.1:5173',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: MICRO ? undefined : [
    {
      command: `go run api/main.go -f "${backendConfig}"`,
      cwd: '../backend',
      url: 'http://127.0.0.1:8080/api/v1/livez',
      reuseExistingServer: true,
      timeout: 60_000,
    },
    {
      // 生产构建 + preview：与部署形态一致，也避免 vite dev 冷缓存拖慢 CI
      command: 'npm run build && npm run preview -- --host 127.0.0.1 --port 5173',
      url: 'http://127.0.0.1:5173',
      reuseExistingServer: true,
      timeout: 240_000,
    },
  ],
})
