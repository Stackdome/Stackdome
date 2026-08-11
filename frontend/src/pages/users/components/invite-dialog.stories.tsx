import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, waitFor, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { makeProject } from '../../../../.storybook/fixtures'
import { baselineHandlers } from '../../../../.storybook/msw-handlers'
import { InviteDialog } from './invite-dialog'

const PROJECTS_PATH = '/api/v1/organizations/:orgId/projects'
const INVITES_PATH = '/api/v1/organizations/:orgId/invites'

const twoProjects = [
  makeProject({ id: 'p1', name: 'default', default_project: true }),
  makeProject({ id: 'p2', name: 'platform', default_project: false }),
]

const withProjects = [
  http.get(PROJECTS_PATH, () => HttpResponse.json({ items: twoProjects, total: twoProjects.length })),
  ...baselineHandlers,
]

const meta = {
  title: 'Features/Users/InviteDialog',
  component: InviteDialog,
  tags: ['ai-generated'],
  args: {
    open: true,
    onOpenChange: fn(),
    onCreated: fn(),
  },
  parameters: { msw: withProjects },
} satisfies Meta<typeof InviteDialog>

export default meta
type Story = StoryObj<typeof meta>

/** Default form — role cards use an ink tint for the selected state, never
 *  brand orange (rubric #2/#3: black/ink is the only "active" colour). */
export const Default: Story = {
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByRole('dialog')).toBeInTheDocument()
    await expect(body.getByLabelText(/email/i)).toBeInTheDocument()

    const developerCard = body.getByText('Developer').closest('label')
    await expect(developerCard).toBeTruthy()
    await expect(developerCard?.className).not.toContain('border-brand')
    await expect(developerCard?.className).toContain('bg-foreground/5')

    const viewerCard = body.getByText('Viewer').closest('label')
    await expect(viewerCard?.className).not.toContain('bg-foreground/5')
  },
}

/** Selecting the other role moves the ink tint — no color, no movement. */
export const RoleSelection: Story = {
  play: async ({ canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    await body.findByRole('dialog')
    const viewerRadio = body.getByRole('radio', { name: /viewer/i })
    await userEvent.click(viewerRadio)
    await expect(viewerRadio).toHaveAttribute('data-state', 'checked')
    const viewerCard = body.getByText('Viewer').closest('label')
    await expect(viewerCard?.className).toContain('bg-foreground/5')
    await expect(viewerCard?.className).not.toContain('text-brand')
  },
}

export const ValidationError: Story = {
  play: async ({ canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    await body.findByRole('dialog')
    await userEvent.type(body.getByLabelText(/email/i), 'not-an-email')
    await userEvent.click(body.getByRole('button', { name: /send invitation/i }))
    await expect(await body.findByText(/valid email/i)).toBeInTheDocument()
  },
}

/** Top-level submission failure renders through the shared AlertBanner
 *  (danger tokens, hairline border) instead of a page-patched banner. */
export const ServerError: Story = {
  parameters: {
    msw: [
      http.post(INVITES_PATH, () => HttpResponse.json({ reason: 'no seats remaining' }, { status: 402 })),
      ...withProjects,
    ],
  },
  play: async ({ canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    await body.findByRole('dialog')
    await userEvent.type(body.getByLabelText(/email/i), 'a@b.io')
    await userEvent.click(body.getByRole('button', { name: /send invitation/i }))
    const alert = await body.findByRole('alert')
    await expect(alert).toHaveTextContent('no seats remaining')
    await expect(alert.className).toContain('border-danger-border')

    // Dismiss clears the banner without touching the form.
    await userEvent.click(body.getByRole('button', { name: /dismiss/i }))
    await expect(body.queryByRole('alert')).toBeNull()
  },
}

/** email_sent=true — one-time link + "SHOWN ONCE" meta label, success tokens. */
export const SuccessSent: Story = {
  parameters: {
    msw: [
      http.post(INVITES_PATH, () =>
        HttpResponse.json({ email_sent: true, invite_token: 'tok_abc123' }),
      ),
      ...withProjects,
    ],
  },
  play: async ({ canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    await body.findByRole('dialog')
    await userEvent.type(body.getByLabelText(/email/i), 'a@b.io')
    await userEvent.click(body.getByRole('button', { name: /send invitation/i }))
    await expect(await body.findByText(/tok_abc123/)).toBeInTheDocument()
    const shownOnce = body.getByText('SHOWN ONCE')
    await expect(shownOnce.className).not.toContain('text-brand')
  },
}

/** email_sent=false — fallback warn banner, link still usable. */
export const SuccessEmailFailed: Story = {
  parameters: {
    msw: [
      http.post(INVITES_PATH, () =>
        HttpResponse.json({
          email_sent: false,
          email_error: 'mailbox unavailable',
          invite_token: 'tok_def456',
        }),
      ),
      ...withProjects,
    ],
  },
  play: async ({ canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    await body.findByRole('dialog')
    await userEvent.type(body.getByLabelText(/email/i), 'a@b.io')
    await userEvent.click(body.getByRole('button', { name: /send invitation/i }))
    await expect(await body.findByText(/email delivery failed/i)).toBeInTheDocument()
    await waitFor(async () => {
      await expect(body.getAllByText(/tok_def456/).length).toBeGreaterThan(0)
    })
  },
}
