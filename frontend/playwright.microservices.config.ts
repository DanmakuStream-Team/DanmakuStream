import path from 'node:path'
import { defineConfig, devices } from '@playwright/test'

// 微服务模式 E2E：
// - 不启动单体 Go/Vite 进程，compose 栈由 scripts/run-microservices-e2e.sh 管理；
// - 用例目录与单体统一为 ./e2e，运行时通过 E2E_MICROSERVICES=1 切换 API/GATEWAY 基址；
// - 可通过 E2E_GATEWAY_URL / E2E_API_BASE / MICRO_E2E_FRONTEND_URL 覆盖默认值。
const frontendURL = process.env.MICRO_E2E_FRONTEND_URL ?? 'http://127.0.0.1:18080'
const artifactRoot = process.env.MICRO_E2E_ARTIFACT_DIR
  ?? path.resolve('../artifacts/microservices-e2e')

process.env.E2E_MICROSERVICES = '1'
process.env.E2E_GATEWAY_URL ??= process.env.MICRO_E2E_GATEWAY_URL ?? 'http://127.0.0.1:18888'
process.env.E2E_API_BASE ??= `${process.env.E2E_GATEWAY_URL}/api/v1`

export default defineConfig({
  testDir: './e2e',
  outputDir: path.join(artifactRoot, 'test-results'),
  globalSetup: './e2e/global-setup.ts',
  globalTeardown: './e2e/global-teardown.ts',
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: [
    ['list'],
    ['html', { outputFolder: path.join(artifactRoot, 'playwright-report'), open: 'never' }],
  ],
  use: {
    baseURL: frontendURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
