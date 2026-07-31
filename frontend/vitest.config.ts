import path from 'path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'
import { playwright } from '@vitest/browser-playwright'
import { storybookTest } from '@storybook/addon-vitest/vitest-plugin'

const dirname = path.dirname(fileURLToPath(import.meta.url))

// Two projects: the jsdom unit suite, and every story run as a test in a real
// browser. The story project renders each story and executes its play
// function, so a component change that breaks a story fails the build instead
// of rotting until someone opens Storybook.
export default defineConfig({
  test: {
    projects: [
      {
        extends: './vite.config.ts',
        test: {
          name: 'unit',
          // Environment stays per-file via the @vitest-environment pragma the
          // existing suites already carry.
          include: ['src/**/*.{test,spec}.{ts,tsx}'],
          sequence: { groupOrder: 0 },
        },
      },
      {
        extends: './vite.config.ts',
        plugins: [storybookTest({ configDir: path.join(dirname, '.storybook') })],
        test: {
          name: 'storybook',
          browser: {
            enabled: true,
            headless: true,
            provider: playwright(),
            instances: [{ browser: 'chromium' }],
          },
          // Runs after the unit project rather than alongside it: a browser
          // project and a 178-file jsdom suite oversubscribe the cores and
          // starve each other into timeouts.
          sequence: { groupOrder: 1 },
        },
      },
    ],
  },
})
