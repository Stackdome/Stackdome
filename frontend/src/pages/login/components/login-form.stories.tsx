import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { baselineHandlers } from '../../../../.storybook/msw-handlers'
import { withCurrentUser } from '../../../../.storybook/decorators'
import { AuthShell, SwapLink } from '@/pages/auth/components/auth-shell'
import { LoginForm } from './login-form'

const CONFIG_PATH = '/api/v1/config'
const LOGIN_PATH = '/api/v1/auth/login'

function withConfig(githubOAuth: boolean) {
  return http.get(CONFIG_PATH, () => HttpResponse.json({ github_oauth: githubOAuth }))
}

/** "Continue" also matches the GitHub button's accessible name ("Continue
 *  with GitHub"), so the email/password submit needs an unambiguous query. */
function getSubmitButton(canvasElement: HTMLElement): HTMLButtonElement {
  const button = canvasElement.querySelector('button[type="submit"]')
  if (!button) throw new Error('submit button not found')
  return button as HTMLButtonElement
}

/** Renders the real page composition (room + plate) without the page's
 *  isUserLoggedIn() redirect guard, which would fire against the session's
 *  seeded auth state before the story ever paints. */
function LoginScreen() {
  return (
    <AuthShell
      title="Welcome back."
      sub="Sign in to manage your stacks."
      below={<SwapLink lead="New to Stackdome?" to="/sign-up" label="Create an account" />}
    >
      <LoginForm />
    </AuthShell>
  )
}

const meta = {
  title: 'Pages/Login',
  component: LoginScreen,
  tags: ['ai-generated'],
  parameters: { layout: 'fullscreen', msw: [withConfig(true), ...baselineHandlers] },
  decorators: [withCurrentUser],
} satisfies Meta<typeof LoginScreen>

export default meta
type Story = StoryObj<typeof meta>

/** GitHub OAuth configured: the outline "Continue with GitHub" sits above the
 *  form. The email/password submit stays the screen's one filled control
 *  (rubric #2 — one primary action, no doubled-up filled pills). */
export const Default: Story = {
  play: async ({ canvas, canvasElement }) => {
    const githubButton = await canvas.findByRole('button', { name: /continue with github/i })
    await expect(githubButton.className).not.toContain('bg-foreground')

    const submit = getSubmitButton(canvasElement)
    await expect(submit.className).toContain('bg-primary')
  },
}

// A "no GitHub OAuth" story is intentionally omitted: getAppConfig() (src/api/config.ts)
// single-flights and caches its result at module scope for the session, so a
// second story's differing /api/v1/config mock in the same file would race
// against the first story's already-resolved value instead of reflecting it.
// components/auth/tests/github-sign-in-button.test.tsx covers the disabled
// (renders null) case with a mocked hook instead, which sidesteps the cache.

/** Empty submit surfaces inline field errors via the branded FieldError,
 *  never a raw ad hoc <p>. */
export const ValidationErrors: Story = {
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(getSubmitButton(canvasElement))
    await expect(await canvas.findByText('Email is invalid')).toBeInTheDocument()
    await expect(await canvas.findByText('Password is required')).toBeInTheDocument()
  },
}

/** Backend rejection renders through AlertBanner, not a hand-rolled danger div. */
export const ServerError: Story = {
  parameters: {
    msw: [
      withConfig(true),
      http.post(LOGIN_PATH, () => HttpResponse.json({ reason: 'Invalid email or password' }, { status: 401 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.type(canvas.getByLabelText('Email'), 'ada@example.com')
    await userEvent.type(canvas.getByLabelText('Password'), 'wrong-password')
    await userEvent.click(getSubmitButton(canvasElement))

    const banner = await canvas.findByRole('alert')
    await expect(banner).toHaveTextContent('Invalid email or password')
  },
}

/** Submitting state: spinner + disabled controls, no layout shift. */
export const Submitting: Story = {
  parameters: {
    msw: [withConfig(true), http.post(LOGIN_PATH, () => new Promise(() => {})), ...baselineHandlers],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.type(canvas.getByLabelText('Email'), 'ada@example.com')
    await userEvent.type(canvas.getByLabelText('Password'), 'correct-horse')
    await userEvent.click(getSubmitButton(canvasElement))

    await waitFor(async () => {
      await expect(canvas.getByLabelText('Email')).toBeDisabled()
      await expect(canvas.getByText(/signing in/i)).toBeInTheDocument()
    })
  },
}
