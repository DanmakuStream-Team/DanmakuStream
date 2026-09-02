import path from 'node:path'
import { defineConfig, devices } from '@playwright/test'

const frontendURL = process.env.MICRO_E2E_FRONTEND_URL ?? 'http://127.0.0.1:18080'
const artifactRoot = process.env.MICRO_E2E_ARTIFACT_DIR
  ?? path.resolve('../artifacts/microservices-e2e')

export default defineConfig({
  testDir: './e2e-microservices',
  outputDir: path.join(artifactRoot, 'test-results-full'),
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
