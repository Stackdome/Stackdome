import { useEffect, useState, type ReactNode } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { Route, Routes, useNavigate } from 'react-router-dom'
import { ORG_ID, DEFAULT_PROJECT } from '../../../.storybook/fixtures'
import { baselineHandlers } from '../../../.storybook/msw-handlers'
import { withConfirm, withCurrentUser } from '../../../.storybook/decorators'
import type { StackPreviewConfig } from '@/api/preview-configs'
import type { PreviewStack } from '@/api/preview-envs'
import PreviewConfigDetailPage from './config-detail'

// The global preview decorator already supplies a MemoryRouter (nesting a
// second one throws); this hops that router to the detail route so
// useParams()/:configId resolves before rendering the page.
function RouteHost({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const [ready, setReady] = useState(false)
  useEffect(() => {
    navigate('/previews/c1', { replace: true })
    setReady(true)
  }, [navigate])
  if (!ready) return null
  return (
    <Routes>
      <Route path="/previews/:configId" element={<>{children}</>} />
    </Routes>
  )
}

const CONFIG_PATH = `/api/v1/organizations/${ORG_ID}/projects/${DEFAULT_PROJECT}/stack-preview-configs/c1`
const ENVS_PATH = `/api/v1/organizations/${ORG_ID}/projects/${DEFAULT_PROJECT}/preview-stacks`

const config: StackPreviewConfig = {
  id: 'c1',
  name: 'webapp',
  git_repository: { repo_url: 'https://github.com/acme/webapp.git', base_branch: 'main' },
  stackfile_path: 'stackfile.yaml',
  max_active_previews: 10,
} as StackPreviewConfig

const mixedEnvs: PreviewStack[] = [
  {
    id: 'e1',
    pr_number: '101',
    branch: 'feat/ready-one',
    commit: 'aaa1111bbbb',
    config_id: 'c1',
    status: { phase: 'Ready' },
  },
  {
    id: 'e2',
    pr_number: '202',
    branch: 'feat/pending-two',
    commit: 'ccc2222dddd',
    config_id: 'c1',
    status: { phase: 'Provisioning' },
  },
  {
    id: 'e3',
    pr_number: '303',
    branch: 'feat/failed-three',
    commit: 'eee3333ffff',
    config_id: 'c1',
    status: { phase: 'Failed', reason: 'BuildFailed', message: 'image build failed' },
  },
] as PreviewStack[]

const meta = {
  title: 'Pages/PreviewConfigDetail',
  component: PreviewConfigDetailPage,
  tags: ['ai-generated'],
  decorators: [
    withConfirm,
    withCurrentUser,
    (Story) => (
      <RouteHost>
        <Story />
      </RouteHost>
    ),
  ],
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof PreviewConfigDetailPage>

export default meta
type Story = StoryObj<typeof meta>

export const Populated: Story = {
  parameters: {
    msw: [
      http.get(CONFIG_PATH, () => HttpResponse.json(config)),
      http.get(ENVS_PATH, () => HttpResponse.json({ items: mixedEnvs, total: mixedEnvs.length })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('PR #101')).toBeInTheDocument()
      await expect(canvas.getByText('PR #202')).toBeInTheDocument()
      await expect(canvas.getByText('PR #303')).toBeInTheDocument()
    })
  },
}

// Status/Sort triggers are the sanctioned Button primitive now, not a
// hand-rolled mono/uppercase `<button>` — Geist casing, never brand orange
// for the selected state (rubric #3, #8; mirrors the fix on the stacks list
// toolbar in stack-card.tsx's sibling page).
export const FilterContract: Story = {
  parameters: {
    msw: [
      http.get(CONFIG_PATH, () => HttpResponse.json(config)),
      http.get(ENVS_PATH, () => HttpResponse.json({ items: mixedEnvs, total: mixedEnvs.length })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('PR #101')).toBeInTheDocument()
    })

    const statusTrigger = canvas.getByRole('button', { name: /status:/i })
    await expect(statusTrigger.className).not.toContain('font-mono')
    await expect(statusTrigger.className).not.toContain('uppercase')

    await userEvent.click(statusTrigger)
    const menu = within(canvasElement.ownerDocument.body)
    const failedOption = await menu.findByRole('menuitem', { name: /failed/i })
    await expect(failedOption.className).not.toContain('font-mono')
    await expect(failedOption.className).not.toContain('text-brand')

    await userEvent.click(failedOption)
    await waitFor(async () => {
      await expect(canvas.getByText('PR #303')).toBeInTheDocument()
      await expect(canvas.queryByText('PR #101')).toBeNull()
    })
  },
}

// The settings modal's "Delete configuration" trigger is the danger-outline
// contract (outline + danger text/border, not a solid destructive fill) —
// the solid `destructive` variant is reserved for the confirm dialog's own
// commit button (mirrors clusters detail's "Delete Cluster" button).
export const SettingsDangerZone: Story = {
  parameters: {
    msw: [
      http.get(CONFIG_PATH, () => HttpResponse.json(config)),
      http.get(ENVS_PATH, () => HttpResponse.json({ items: mixedEnvs, total: mixedEnvs.length })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('PR #101')).toBeInTheDocument()
    })

    await userEvent.click(canvas.getByRole('button', { name: /^settings$/i }));
    // Dialog content portals to document.body, outside the canvas root.
    const body = within(canvasElement.ownerDocument.body)
    const deleteButton = await body.findByRole('button', { name: /delete configuration/i })
    await expect(deleteButton.className).toContain('border-danger-border')
    await expect(deleteButton.className).toContain('text-danger')
    await expect(deleteButton.className).not.toContain('text-white')
  },
}

export const NoEnvironmentsYet: Story = {
  parameters: {
    msw: [
      http.get(CONFIG_PATH, () => HttpResponse.json(config)),
      http.get(ENVS_PATH, () => HttpResponse.json({ items: [], total: 0 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('No preview environments yet')).toBeInTheDocument()
  },
}

export const ConfigLoadError: Story = {
  parameters: {
    msw: [
      http.get(CONFIG_PATH, () => HttpResponse.json({ message: 'not found' }, { status: 404 })),
      http.get(ENVS_PATH, () => HttpResponse.json({ items: [], total: 0 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText("Couldn't load this configuration")).toBeInTheDocument()
  },
}
