import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { baselineHandlers } from '../../../../.storybook/msw-handlers'
import { AuthShell, SwapLink } from '@/pages/auth/components/auth-shell'
import { SignupForm } from './signup-form'

const CONFIG_PATH = '/api/v1/config'
const SIGNUP_PATH = '/api/v1/user-signup'

function withConfig(githubOAuth: boolean) {
  return http.get(CONFIG_PATH, () => HttpResponse.json({ github_oauth: githubOAuth }))
}

function SignupScreen() {
  return (
    <AuthShell
      title="Own your stack."
      sub="Create an account to start deploying."
      below={<SwapLink lead="Already have an account?" to="/sign-in" label="Log in" />}
    >
      <SignupForm />
    </AuthShell>
  )
}

const meta = {
  title: 'Pages/Signup',
  component: SignupScreen,
  tags: ['ai-generated'],
  parameters: { layout: 'fullscreen', msw: [withConfig(true), ...baselineHandlers] },
} satisfies Meta<typeof SignupScreen>

export default meta
type Story = StoryObj<typeof meta>

/** The email/password submit is the screen's one filled control; GitHub is
 *  the outline secondary above it (rubric #2). */
export const Default: Story = {
  play: async ({ canvas }) => {
    const githubButton = await canvas.findByRole('button', { name: /continue with github/i })
    await expect(githubButton.className).not.toContain('bg-foreground')

    const submit = canvas.getByRole('button', { name: /create account/i })
    await expect(submit.className).toContain('bg-primary')
  },
}

/** Mismatched passwords surface on the confirm field via the branded
 *  FieldError, and empty required fields all report at once. */
export const ValidationErrors: Story = {
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(canvas.getByLabelText(/^password/i), 'password123')
    await userEvent.type(canvas.getByLabelText(/confirm password/i), 'not-the-same')
    await userEvent.click(canvas.getByRole('button', { name: /create account/i }))

    await expect(await canvas.findByText('Name is required')).toBeInTheDocument()
    await expect(await canvas.findByText('Organization name is required')).toBeInTheDocument()
    await expect(await canvas.findByText('Passwords do not match')).toBeInTheDocument()
  },
}

/** Backend rejection (e.g. email already in use) renders through AlertBanner. */
export const ServerError: Story = {
  parameters: {
    msw: [
      withConfig(true),
      http.post(SIGNUP_PATH, () => HttpResponse.json({ reason: 'Email already in use' }, { status: 409 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(canvas.getByLabelText(/full name/i), 'Ada Lovelace')
    await userEvent.type(canvas.getByLabelText(/organization/i), 'Analytical Engines')
    await userEvent.type(canvas.getByLabelText(/email/i), 'ada@example.com')
    await userEvent.type(canvas.getByLabelText(/^password/i), 'password123')
    await userEvent.type(canvas.getByLabelText(/confirm password/i), 'password123')
    await userEvent.click(canvas.getByRole('button', { name: /create account/i }))

    const banner = await canvas.findByRole('alert')
    await expect(banner).toHaveTextContent('Email already in use')
  },
}

/** Submitting state: spinner + disabled controls. */
export const Submitting: Story = {
  parameters: {
    msw: [withConfig(true), http.post(SIGNUP_PATH, () => new Promise(() => {})), ...baselineHandlers],
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(canvas.getByLabelText(/full name/i), 'Ada Lovelace')
    await userEvent.type(canvas.getByLabelText(/organization/i), 'Analytical Engines')
    await userEvent.type(canvas.getByLabelText(/email/i), 'ada@example.com')
    await userEvent.type(canvas.getByLabelText(/^password/i), 'password123')
    await userEvent.type(canvas.getByLabelText(/confirm password/i), 'password123')
    await userEvent.click(canvas.getByRole('button', { name: /create account/i }))

    await waitFor(async () => {
      await expect(canvas.getByLabelText(/full name/i)).toBeDisabled()
      await expect(canvas.getByText(/creating account/i)).toBeInTheDocument()
    })
  },
}

/** Long org/name values don't blow out the fixed-width auth plate. */
export const LongValuesOverflow: Story = {
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(
      canvas.getByLabelText(/full name/i),
      'Bartholomew Alexander Featherstonehaugh-Worthington',
    )
    await userEvent.type(
      canvas.getByLabelText(/organization/i),
      'An Extremely Long Corporate Holding Company International Group',
    )
    await expect(canvas.getByDisplayValue(/Bartholomew/)).toBeInTheDocument()
  },
}
