import path from 'node:path'
import { defineConfig, devices } from '@playwright/test'

// This suite never starts the monolith. The Compose stack is owned by
// scripts/run-microservices-e2e.sh so failures can retain all container logs.
const frontendURL = process.env.MICRO_E2E_FRONTEND_URL ?? 'http://127.0.0.1:18080'
const artifactRoot = process.env.MICRO_E2E_ARTIFACT_DIR
  ?? path.resolve('../artifacts/microservices-e2e')

export default defineConfig({
  testDir: './e2e-microservices',
  outputDir: path.join(artifactRoot, 'test-results'),
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
