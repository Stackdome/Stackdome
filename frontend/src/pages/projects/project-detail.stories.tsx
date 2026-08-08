import { useEffect, useState, type ReactNode } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { Route, Routes, useNavigate } from 'react-router-dom'
import { makeProject, makeProjectMembership, makeUser } from '../../../.storybook/fixtures'
import { baselineHandlers } from '../../../.storybook/msw-handlers'
import { withSheetHeader } from '../../../.storybook/decorators'
import ProjectDetailPage from './project-detail'

// The global preview decorator already supplies a MemoryRouter (nesting a
// second one throws); this hops that router to the detail route so
// useParams()/:projectName resolves before rendering the page.
function RouteHost({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const [ready, setReady] = useState(false)
  useEffect(() => {
    navigate('/settings/projects/platform', { replace: true })
    setReady(true)
  }, [navigate])
  if (!ready) return null
  return (
    <Routes>
      <Route path="/settings/projects/:projectName" element={<>{children}</>} />
    </Routes>
  )
}

const PROJECT_PATH = '/api/v1/organizations/:orgId/projects/:projectName'
const MEMBERS_PATH = `${PROJECT_PATH}/members`

const project = makeProject({ id: 'p2', name: 'platform', default_project: false, created_at: '2026-06-15T09:00:00Z' })

const members = [
  makeProjectMembership({
    id: 'pm-1',
    role: 'Developer',
    user: makeUser({ id: 'u1', name: 'Ada Lovelace', email: 'ada@example.com', role: 'OrgAdmin' }),
  }),
  makeProjectMembership({
    id: 'pm-2',
    role: 'Viewer',
    user_id: 'u2',
    user: makeUser({ id: 'u2', name: 'Grace Hopper', email: 'grace@example.com', role: 'OrgMember' }),
  }),
]

const meta = {
  title: 'Pages/ProjectDetail',
  component: ProjectDetailPage,
  tags: ['ai-generated'],
  decorators: [
    withSheetHeader,
    (Story) => (
      <RouteHost>
        <Story />
      </RouteHost>
    ),
  ],
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof ProjectDetailPage>

export default meta
type Story = StoryObj<typeof meta>

export const Populated: Story = {
  parameters: {
    msw: [
      http.get(PROJECT_PATH, () => HttpResponse.json(project)),
      http.get(MEMBERS_PATH, () => HttpResponse.json({ items: members, total: members.length })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('Ada Lovelace')).toBeInTheDocument()
      await expect(canvas.getByText('Grace Hopper')).toBeInTheDocument()
    })
  },
}

export const NoMembersYet: Story = {
  parameters: {
    msw: [
      http.get(PROJECT_PATH, () => HttpResponse.json(project)),
      http.get(MEMBERS_PATH, () => HttpResponse.json({ items: [], total: 0 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('No members yet')).toBeInTheDocument()
  },
}

// Search narrows to zero matches — a distinct, non-dashed empty state from
// the "add your first member" case (rubric watch-list carries over from the
// list-page loop: empty states must read differently for "nothing yet" vs.
// "nothing matched").
export const NoSearchMatch: Story = {
  parameters: {
    msw: [
      http.get(PROJECT_PATH, () => HttpResponse.json(project)),
      http.get(MEMBERS_PATH, () => HttpResponse.json({ items: members, total: members.length })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, userEvent }) => {
    const search = await canvas.findByPlaceholderText('Search members…')
    await userEvent.type(search, 'nobody-matches-this')
    await expect(await canvas.findByText('No members match')).toBeInTheDocument()
  },
}

export const MembersError: Story = {
  parameters: {
    msw: [
      http.get(PROJECT_PATH, () => HttpResponse.json(project)),
      http.get(MEMBERS_PATH, () => HttpResponse.json({ message: 'membership service unreachable' }, { status: 500 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText("Couldn't load members")).toBeInTheDocument()
  },
}

export const DefaultProjectReadOnly: Story = {
  parameters: {
    msw: [
      http.get(PROJECT_PATH, () => HttpResponse.json(makeProject({ name: 'platform', default_project: true }))),
      http.get(MEMBERS_PATH, () => HttpResponse.json({ items: members, total: members.length })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    const renameButton = await canvas.findByRole('button', { name: /rename/i })
    await expect(renameButton).toBeDisabled()
  },
}
