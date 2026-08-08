import { useEffect, useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { useNavigate } from 'react-router-dom'
import { http, HttpResponse } from 'msw'
import { makeStack } from '../../../.storybook/fixtures'
import { baselineHandlers } from '../../../.storybook/msw-handlers'
import { withConfirm, withCurrentUser, withStack } from '../../../.storybook/decorators'
import { ReleaseState } from '@/pages/stacks/components/editor/tabs/deployments/release-states'
import type { Stack } from '@/api/stack-types'
import { AppLayout } from '@/components/app-layout'
import StacksPage from '@/pages/stacks/components/list'

/* The whole platform in one frame — real sidebar, real topnav, real page.
 * Design decisions get made against this, never against a component alone.
 * Exploration surface only: no play functions, nothing asserted. */

const previewEnvsEmpty = http.get(
  '/api/v1/organizations/:orgId/projects/:projectName/preview-stacks',
  () => HttpResponse.json({ items: [], total: 0 }),
)

const stacks = [
  makeStack({
    latest_release: { id: 'r1', state: ReleaseState.Released },
    converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
  } as Partial<Stack>),
  makeStack({
    id: 's2',
    name: 'billing-worker',
    latest_release: { id: 'r2', state: ReleaseState.InProgress },
  } as Partial<Stack>),
  makeStack({
    id: 's3',
    name: 'docs-site',
    latest_release: { id: 'r3', state: ReleaseState.Failed },
  } as Partial<Stack>),
  makeStack({ id: 's4', name: 'staging-sandbox' }),
  makeStack({ id: 's5', name: 'auth-gateway' }),
  makeStack({ id: 's6', name: 'search-indexer' }),
  makeStack({ id: 's7', name: 'notifications' }),
  makeStack({ id: 's8', name: 'admin-console' }),
]

/* The preview decorator already mounts a MemoryRouter; hop it to /stacks so the
 * sidebar's active item and the breadcrumb both resolve before first paint. */
function ShellHarness() {
  const navigate = useNavigate()
  const [ready, setReady] = useState(false)
  useEffect(() => {
    navigate('/stacks', { replace: true })
    setReady(true)
  }, [navigate])
  if (!ready) return null
  return (
    <AppLayout>
      <StacksPage />
    </AppLayout>
  )
}

const meta = {
  title: 'Shell/Platform',
  component: ShellHarness,
  decorators: [withConfirm, withCurrentUser, withStack],
  parameters: {
    layout: 'fullscreen',
    msw: [
      previewEnvsEmpty,
      http.get('/api/v1/organizations/:orgId/stacks', () =>
        HttpResponse.json({ items: stacks, total: stacks.length }),
      ),
      ...baselineHandlers,
    ],
  },
} satisfies Meta<typeof ShellHarness>

export default meta
type Story = StoryObj<typeof meta>

export const StacksList: Story = {}
