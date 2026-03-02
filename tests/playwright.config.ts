import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [['html'], ['list']],
  snapshotDir: './snapshots',
  
  projects: [
    {
      name: 'visual',
      testMatch: /visual\.spec\.ts/,
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3000',
        viewport: { width: 1920, height: 1080 },
      },
    },
    {
      name: 'functional',
      testMatch: /functional\.spec\.ts/,
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3000',
        viewport: { width: 1920, height: 1080 },
      },
    },
  ],

  webServer: {
    command: 'echo "Server should be running"',
    url: 'http://localhost:3000',
    reuseExistingServer: true,
  },
});
