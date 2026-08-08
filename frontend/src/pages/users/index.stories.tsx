import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { makeUser, makeOrgInvite, makeProject } from '../../../.storybook/fixtures'
import { baselineHandlers } from '../../../.storybook/msw-handlers'
import { withConfirm } from '../../../.storybook/decorators'
import UsersPage from './index'

const USERS_PATH = '/api/v1/organizations/:orgId/users'
const INVITES_PATH = '/api/v1/organizations/:orgId/invites'
const PROJECTS_PATH = '/api/v1/organizations/:orgId/projects'

const twoProjects = [
  makeProject({ id: 'p1', name: 'default', default_project: true }),
  makeProject({ id: 'p2', name: 'platform', default_project: false }),
]

const admin = makeUser({
  id: 'u1',
  name: 'Ada Lovelace',
  email: 'ada@example.com',
  role: 'OrgAdmin',
  projects: [{ project_name: 'default', role: 'Developer', default_project: true }],
})
const member = makeUser({
  id: 'u2',
  name: 'Grace Hopper',
  email: 'grace@example.com',
  role: 'OrgMember',
  projects: [
    { project_name: 'default', role: 'Developer', default_project: true },
    { project_name: 'platform', role: 'Viewer', default_project: false },
  ],
})

const pendingInvite = makeOrgInvite({
  id: 'inv-1',
  email: 'newhire@example.com',
  project_name: 'platform',
  role: 'Developer',
  invited_by: 'ada@example.com',
})

function withUsersAndInvites(users: ReturnType<typeof makeUser>[], invites: ReturnType<typeof makeOrgInvite>[]) {
  return [
    http.get(USERS_PATH, () => HttpResponse.json({ items: users, total: users.length })),
    http.get(INVITES_PATH, () => HttpResponse.json({ items: invites, total: invites.length })),
    http.get(PROJECTS_PATH, () => HttpResponse.json({ items: twoProjects, total: twoProjects.length })),
    ...baselineHandlers,
  ]
}

const meta = {
  title: 'Pages/Users',
  component: UsersPage,
  tags: ['ai-generated'],
  parameters: { layout: 'fullscreen' },
  decorators: [withConfirm],
} satisfies Meta<typeof UsersPage>

export default meta
type Story = StoryObj<typeof meta>

/** Mixed active (admin + member) and pending rows — org-role and "invited"
 *  state both render through the shared Badge primitive, never an ad hoc
 *  brand-orange or hand-rolled chip (rubric #3/#4). */
export const Populated: Story = {
  parameters: { msw: withUsersAndInvites([admin, member], [pendingInvite]) },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('Ada Lovelace')).toBeInTheDocument()
      await expect(canvas.getByText('Grace Hopper')).toBeInTheDocument()
      await expect(canvas.getByText('newhire@example.com')).toBeInTheDocument()
    })

    // Org-role badges read through the Badge primitive (no brand-orange span).
    const adminBadge = canvas.getByText('OrgAdmin')
    await expect(adminBadge.className).not.toContain('text-brand')
    await expect(adminBadge.closest('[data-slot="badge"]')).toBeTruthy()

    // "invited" state uses Badge variant="warning" — hue reports real state.
    const invitedBadge = canvas.getByText('invited')
    await expect(invitedBadge.closest('[data-slot="badge"]')).toBeTruthy()

    // Default-project indicator is ink-neutral, not brand orange.
    const defaultTags = canvas.getAllByText('default', { exact: false })
    for (const tag of defaultTags) {
      await expect(tag.className).not.toContain('text-brand')
    }

    // Kebab menu on the admin row opens with a portal-rendered menu.
    const menuButtons = canvas.getAllByRole('button', { name: /user actions/i })
    await userEvent.click(menuButtons[0])
    const menu = within(canvasElement.ownerDocument.body)
    await expect(await menu.findByText('Demote')).toBeInTheDocument()
  },
}

export const Loading: Story = {
  parameters: {
    msw: [
      http.get(USERS_PATH, () => new Promise(() => {})),
      http.get(INVITES_PATH, () => new Promise(() => {})),
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
      http.get(USERS_PATH, () => HttpResponse.json({ reason: 'organization service unreachable' }, { status: 500 })),
      http.get(INVITES_PATH, () => HttpResponse.json({ items: [] })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText("Couldn't load users")).toBeInTheDocument()
    await expect(await canvas.findByRole('button', { name: /retry/i })).toBeInTheDocument()
  },
}

export const Empty: Story = {
  parameters: { msw: withUsersAndInvites([], []) },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('No users yet')).toBeInTheDocument()
    // Header action + empty-state action both render "Invite user".
    await expect(canvas.getAllByRole('button', { name: /invite user/i }).length).toBe(2)
  },
}

/** Filters that match nothing still show the toolbar (rubric: flat, no
 *  full-page empty state when rows exist but are filtered out). */
export const NoSearchMatch: Story = {
  parameters: { msw: withUsersAndInvites([admin, member], [pendingInvite]) },
  play: async ({ canvas, userEvent }) => {
    await waitFor(async () => {
      await expect(canvas.getByText('Ada Lovelace')).toBeInTheDocument()
    })
    await userEvent.type(canvas.getByPlaceholderText(/search by name or email/i), 'nobody-matches-this')
    await expect(await canvas.findByText('No users match these filters')).toBeInTheDocument()
    await userEvent.click(canvas.getByRole('button', { name: /clear filters/i }))
    await expect(await canvas.findByText('Ada Lovelace')).toBeInTheDocument()
  },
}

/** Long name/email truncates in the row instead of wrapping or overflowing. */
export const LongNameOverflow: Story = {
  parameters: {
    msw: withUsersAndInvites(
      [
        makeUser({
          id: 'u3',
          name: 'Bartholomew Alexander Featherstonehaugh-Worthington',
          email: 'bartholomew.alexander.featherstonehaugh-worthington@an-extremely-long-corporate-domain-example.com',
          role: 'OrgMember',
          projects: [{ project_name: 'default', role: 'Developer', default_project: true }],
        }),
      ],
      [],
    ),
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText(/Bartholomew/)).toBeInTheDocument()
  },
}
