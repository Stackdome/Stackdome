import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { makeProject } from '../../../.storybook/fixtures'
import { baselineHandlers } from '../../../.storybook/msw-handlers'
import ProjectsPage from './index'

const PROJECTS_PATH = '/api/v1/organizations/:orgId/projects'

const threeProjects = [
  makeProject({ id: 'p1', name: 'default', default_project: true, created_at: '2026-06-01T09:00:00Z' }),
  makeProject({ id: 'p2', name: 'platform', default_project: false, created_at: '2026-06-15T09:00:00Z' }),
  makeProject({ id: 'p3', name: 'growth', default_project: false, created_at: '2026-07-02T09:00:00Z' }),
]

const meta = {
  title: 'Pages/Projects',
  component: ProjectsPage,
  tags: ['ai-generated'],
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof ProjectsPage>

export default meta
type Story = StoryObj<typeof meta>

export const Populated: Story = {
  parameters: {
    msw: [
      http.get(PROJECTS_PATH, () => HttpResponse.json({ items: threeProjects, total: threeProjects.length })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('platform')).toBeInTheDocument()
      await expect(canvas.getByText('growth')).toBeInTheDocument()
    })

    // Row actions open a portal menu; Delete reads through the shared
    // destructive variant (danger-tinted focus wash), never a hand-rolled
    // text-only color override (rubric #4 — hue reports state).
    const rows = canvas.getAllByRole('button', { name: /project actions/i })
    await userEvent.click(rows[1])
    const menu = within(canvasElement.ownerDocument.body)
    const deleteItem = await menu.findByText('Delete')
    await expect(deleteItem.closest('[data-variant="destructive"]')).toBeTruthy()
  },
}

// Every project is the org default — the "create your first project" empty
// state renders alongside the single-row table, not in place of it.
export const OnlyDefault: Story = {
  parameters: {
    msw: [
      http.get(PROJECTS_PATH, () =>
        HttpResponse.json({ items: [threeProjects[0]], total: 1 }),
      ),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('No additional projects')).toBeInTheDocument()
  },
}

export const Loading: Story = {
  parameters: {
    msw: [
      http.get(PROJECTS_PATH, () => new Promise(() => {})),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0)
  },
}

export const ErrorState: Story = {
  parameters: {
    msw: [
      http.get(PROJECTS_PATH, () => HttpResponse.json({ message: 'organization service unreachable' }, { status: 500 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText("Couldn't load projects")).toBeInTheDocument()
    await expect(await canvas.findByRole('button', { name: /retry/i })).toBeInTheDocument()
  },
}

export const LongName: Story = {
  parameters: {
    msw: [
      http.get(PROJECTS_PATH, () =>
        HttpResponse.json({
          items: [
            threeProjects[0],
            makeProject({
              id: 'p2',
              name: 'an-extremely-long-project-name-that-should-truncate-in-the-table-cell',
              default_project: false,
            }),
          ],
          total: 2,
        }),
      ),
      ...baselineHandlers,
    ],
  },
}
