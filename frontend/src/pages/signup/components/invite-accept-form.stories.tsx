import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { baselineHandlers } from '../../../../.storybook/msw-handlers'
import { AuthShell } from '@/pages/auth/components/auth-shell'
import { InviteAcceptForm } from './invite-accept-form'
import type { OrgInviteInfo } from '@/api/invites'

const CONFIG_PATH = '/api/v1/config'
const SIGNUP_PATH = '/api/v1/user-signup'

function withConfig(githubOAuth: boolean) {
  return http.get(CONFIG_PATH, () => HttpResponse.json({ github_oauth: githubOAuth }))
}

const info = {
  org_name: 'Acme',
  project_name: 'engineering',
  inviter_name: 'Jane',
  expires_at: '2026-05-19T00:00:00Z',
} as OrgInviteInfo

function InviteScreen() {
  return (
    <AuthShell title={`Join ${info.org_name}.`} sub={`${info.inviter_name} invited you to the ${info.project_name} project`}>
      <InviteAcceptForm token="tok_1" info={info} />
    </AuthShell>
  )
}

const meta = {
  title: 'Pages/Signup/Invite',
  component: InviteScreen,
  tags: ['ai-generated'],
  parameters: { layout: 'fullscreen', msw: [withConfig(true), ...baselineHandlers] },
} satisfies Meta<typeof InviteScreen>

export default meta
type Story = StoryObj<typeof meta>

/** Same one-filled-control contract as the standalone signup screen. */
export const InviteForm: Story = {
  play: async ({ canvas }) => {
    const githubButton = await canvas.findByRole('button', { name: /continue with github/i })
    await expect(githubButton.className).not.toContain('bg-foreground')

    const submit = canvas.getByRole('button', { name: /create account and join/i })
    await expect(submit.className).toContain('bg-primary')
  },
}

export const ValidationErrors: Story = {
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: /create account and join/i }))
    await expect(await canvas.findByText('Name is required')).toBeInTheDocument()
    await expect(await canvas.findByText('Email is invalid')).toBeInTheDocument()
    await expect(await canvas.findByText('Password must be at least 8 characters')).toBeInTheDocument()
  },
}

/** An invite already claimed by an existing account — the form gives way to
 *  a plain sign-in link instead of retrying signup. */
export const ExistingUser: Story = {
  parameters: {
    msw: [
      withConfig(true),
      http.post(SIGNUP_PATH, () => HttpResponse.json({ reason: 'account exists' }, { status: 409 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(canvas.getByLabelText(/full name/i), 'Ada Lovelace')
    await userEvent.type(canvas.getByLabelText(/^email/i), 'ada@example.com')
    await userEvent.type(canvas.getByLabelText(/^password/i), 'password123')
    await userEvent.click(canvas.getByRole('button', { name: /create account and join/i }))

    await expect(await canvas.findByText(/you already have an account/i)).toBeInTheDocument()
    await expect(canvas.getByRole('link', { name: /log in/i })).toBeInTheDocument()
  },
}

/** Accepting phase: spinner + disabled fields, no editing mid-request. */
export const Accepting: Story = {
  parameters: {
    msw: [withConfig(true), http.post(SIGNUP_PATH, () => new Promise(() => {})), ...baselineHandlers],
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(canvas.getByLabelText(/full name/i), 'Ada Lovelace')
    await userEvent.type(canvas.getByLabelText(/^email/i), 'ada@example.com')
    await userEvent.type(canvas.getByLabelText(/^password/i), 'password123')
    await userEvent.click(canvas.getByRole('button', { name: /create account and join/i }))

    await waitFor(async () => {
      await expect(canvas.getByLabelText(/full name/i)).toBeDisabled()
      await expect(canvas.getByText(/creating account/i)).toBeInTheDocument()
    })
  },
}

/** Successful accept swaps the form for a short confirmation before the
 *  redirect timer fires. */
export const Accepted: Story = {
  parameters: {
    msw: [
      withConfig(true),
      http.post(SIGNUP_PATH, () =>
        HttpResponse.json({ jwt_token: 'tok', user: { id: 'u1', email: 'ada@example.com' } }),
      ),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(canvas.getByLabelText(/full name/i), 'Ada Lovelace')
    await userEvent.type(canvas.getByLabelText(/^email/i), 'ada@example.com')
    await userEvent.type(canvas.getByLabelText(/^password/i), 'password123')
    await userEvent.click(canvas.getByRole('button', { name: /create account and join/i }))

    await waitFor(async () => {
      await expect(await canvas.findByText(/you're in/i)).toBeInTheDocument()
    })
  },
}
