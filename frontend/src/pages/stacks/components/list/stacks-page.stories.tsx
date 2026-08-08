import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { makeStack } from '../../../../../.storybook/fixtures'
import { baselineHandlers } from '../../../../../.storybook/msw-handlers'
import { withConfirm, withCurrentUser, withStack } from '../../../../../.storybook/decorators'
import { ReleaseState } from '@/pages/stacks/components/editor/tabs/deployments/release-states'
import type { Stack } from '@/api/stack-types'
import StacksPage from './index'

const previewEnvsEmpty = http.get(
  '/api/v1/organizations/:orgId/projects/:projectName/preview-stacks',
  () => HttpResponse.json({ items: [], total: 0 }),
)

const mixedStacks = [
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
]

const meta = {
  title: 'Pages/Stacks',
  component: StacksPage,
  tags: ['ai-generated'],
  decorators: [withConfirm, withCurrentUser, withStack],
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof StacksPage>

export default meta
type Story = StoryObj<typeof meta>

export const Populated: Story = {
  parameters: {
    msw: [
      previewEnvsEmpty,
      http.get('/api/v1/organizations/:orgId/stacks', () =>
        HttpResponse.json({ items: mixedStacks, total: mixedStacks.length }),
      ),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('billing-worker')).toBeInTheDocument()
      await expect(canvas.getByText('docs-site')).toBeInTheDocument()
    }, { timeout: 5000 })

    // Status filter is a functional control — Geist, not mono/uppercase
    // chrome, and never brand orange for the selected state (rubric #3, #8).
    await userEvent.click(canvas.getByRole('button', { name: /status/i }))
    const menu = within(canvasElement.ownerDocument.body)
    const okOption = await menu.findByRole('menuitem', { name: /^ok/i })
    await expect(okOption.className).not.toContain('font-mono')
    await expect(okOption.className).not.toContain('text-brand')
    await userEvent.click(okOption)
    await waitFor(async () => {
      await expect(canvas.getByText('orders-api')).toBeInTheDocument()
      await expect(canvas.queryByText('billing-worker')).toBeNull()
      await expect(canvas.queryByText('docs-site')).toBeNull()
    })
  },
}

export const Empty: Story = {
  parameters: {
    msw: [
      previewEnvsEmpty,
      http.get('/api/v1/organizations/:orgId/stacks', () =>
        HttpResponse.json({ items: [], total: 0 }),
      ),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('No stacks deployed yet')).toBeInTheDocument()
  },
}

export const Error: Story = {
  parameters: {
    msw: [
      previewEnvsEmpty,
      http.get('/api/v1/organizations/:orgId/stacks', () =>
        HttpResponse.json({ message: 'cluster hub unreachable' }, { status: 500 }),
      ),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByRole('button', { name: /try again/i })).toBeInTheDocument()
  },
}
