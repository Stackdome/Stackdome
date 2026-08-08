import { makeUser } from '../../.storybook/fixtures'
import { previewHandlers } from './handlers'

/**
 * Boots the browser preview (`pnpm dev:mock`): the **real** app — real router,
 * real shell, real pages — against a mocked network.
 *
 * No Go server, no Postgres, no Kind cluster. Storybook judges components;
 * this judges the product, in a browser, with navigation that actually works.
 */
export async function startPreview(): Promise<void> {
  // Auth is localStorage-based (`src/lib/common.ts`), so signing in is just
  // seeding it. Without this the router bounces straight to /sign-in.
  const user = makeUser()
  localStorage.setItem('authToken', 'preview-token')
  localStorage.setItem('refreshToken', 'preview-refresh')
  localStorage.setItem('currentUser', JSON.stringify(user))

  const { setupWorker } = await import('msw/browser')
  const worker = setupWorker(...previewHandlers)

  // `onUnhandledRequest: 'bypass'` keeps Vite's own asset and HMR traffic out
  // of the console — only /api/v1 is mocked, and that has a catch-all.
  await worker.start({
    onUnhandledRequest: 'bypass',
    quiet: true,
    serviceWorker: { url: '/mockServiceWorker.js' },
  })

  // eslint-disable-next-line no-console
  console.info(
    `%c preview `,
    'background:#191714;color:#F5F4F1;border-radius:4px',
    'mocked network — no backend required. Signed in as ' + user.name,
  )
}
