import path from 'node:path'
import { defineConfig, devices } from '@playwright/test'

// 微服务全量套件：
// - 与 playwright.microservices.config.ts 使用同一套 e2e 目录；
// - 额外产出 junit，以及更长的 navigationTimeout；
// - 仍通过 E2E_MICROSERVICES=1 在运行期切换网关基址。
const frontendURL = process.env.MICRO_E2E_FRONTEND_URL ?? 'http://127.0.0.1:18080'
const artifactRoot = process.env.MICRO_E2E_ARTIFACT_DIR
  ?? path.resolve('../artifacts/microservices-e2e')

process.env.E2E_MICROSERVICES = '1'
process.env.E2E_GATEWAY_URL ??= process.env.MICRO_E2E_GATEWAY_URL ?? 'http://127.0.0.1:18888'
process.env.E2E_API_BASE ??= `${process.env.E2E_GATEWAY_URL}/api/v1`

export default defineConfig({
  testDir: './e2e',
  outputDir: path.join(artifactRoot, 'test-results-full'),
  globalSetup: './e2e/global-setup.ts',
  globalTeardown: './e2e/global-teardown.ts',
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: [
    ['list'],
    ['junit', { outputFile: path.join(artifactRoot, 'junit-microservices-full.xml') }],
    ['html', { outputFolder: path.join(artifactRoot, 'playwright-report-full'), open: 'never' }],
  ],
  use: {
    baseURL: frontendURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 15_000,
    navigationTimeout: 45_000,
  },
  projects: [{ name: 'chromium-micro', use: { ...devices['Desktop Chrome'] } }],
})
