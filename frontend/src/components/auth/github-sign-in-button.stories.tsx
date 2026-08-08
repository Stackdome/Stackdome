import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { baselineHandlers } from '../../../.storybook/msw-handlers'
import { GitHubSignInButton } from './github-sign-in-button'

const CONFIG_PATH = '/api/v1/config'

const meta = {
  title: 'Features/Auth/GitHubSignInButton',
  component: GitHubSignInButton,
  tags: ['ai-generated'],
  parameters: {
    msw: [http.get(CONFIG_PATH, () => HttpResponse.json({ github_oauth: true })), ...baselineHandlers],
  },
} satisfies Meta<typeof GitHubSignInButton>

export default meta
type Story = StoryObj<typeof meta>

// The disabled (renders null when GitHub OAuth isn't configured) branch is
// covered by components/auth/tests/github-sign-in-button.test.tsx instead —
// getAppConfig() single-flights and caches per session, so a second story in
// this file with a different /api/v1/config mock would race the first.
/** Outline secondary, not the filled control — the email/password submit
 *  below it owns that role (rubric #2, one filled control per screen). */
export const Enabled: Story = {
  play: async ({ canvas }) => {
    const button = await canvas.findByRole('button', { name: /continue with github/i })
    await expect(button.className).toContain('bg-control')
    await expect(button.className).not.toContain('bg-foreground')
  },
}
