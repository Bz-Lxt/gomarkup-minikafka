import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  testMatch: 'e2e_flow.spec.ts',
  timeout: 30000,
  use: {
    baseURL: process.env.BASE_URL || 'http://127.0.0.1:18591',
    headless: true,
  },
})
