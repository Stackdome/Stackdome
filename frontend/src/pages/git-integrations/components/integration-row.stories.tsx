import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { IntegrationRow } from './integration-row'
import { GIT_INTEGRATION_TYPE_GITHUB_APP } from '@/lib/git-integrations'
import { ORG_ID } from '../../../../.storybook/fixtures'
import type { GitIntegration } from '@/api/git-integrations'

const INSTALLATIONS_URL = `/api/v1/organizations/${ORG_ID}/git-integrations/:id/installations`

function makeIntegration(overrides: Partial<GitIntegration> = {}): GitIntegration {
  return {
    id: 'gi1',
    type: 'git_credentials',
    host: 'gitlab.com',
    status: 'active',
    credentials_configured: true,
    ...overrides,
  }
}

const meta = {
  title: 'Features/GitIntegrations/IntegrationRow',
  component: IntegrationRow,
  tags: ['ai-generated'],
  args: { onVerify: fn(), onRemove: fn(), onUpdateCredentials: fn() },
  parameters: {
    msw: [http.get(INSTALLATIONS_URL, () => HttpResponse.json({ items: [] }))],
  },
  decorators: [
    (Story) => (
      <div className="max-w-[900px] divide-y divide-border rounded-md border">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof IntegrationRow>

export default meta
type Story = StoryObj<typeof meta>

export const Connected: Story = {
  args: { integration: makeIntegration() },
}

// credentials_configured: false — the row's "action needed" banner. Its CTA
// is a real <Button>, so it must stay ink (rubric #2/#3), never the
// brand-orange chip this used to render as.
export const ActionNeeded: Story = {
  args: { integration: makeIntegration({ credentials_configured: false }) },
  play: async ({ canvas }) => {
    const cta = canvas.getByRole('button', { name: 'Update credentials →' })
    await expect(cta.className).not.toContain('text-brand')
    await expect(cta.className).not.toContain('border-brand')
  },
}

// pending_install github_app row — the banner's CTA is a real link out to
// GitHub, so per contract it reads as an ink underline link, not a chip.
export const NeedsSetup: Story = {
  args: {
    integration: makeIntegration({
      id: 'gi2',
      type: GIT_INTEGRATION_TYPE_GITHUB_APP,
      host: 'github.com',
      status: 'pending_install',
      credentials_configured: true,
      install_url: 'https://github.com/apps/stackdome/installations/new',
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const link = await canvas.findByRole('link', { name: 'Finish install →' })
    await expect(link.className).toContain('underline-offset-2')
    await expect(link.className).not.toContain('text-brand')
    await expect(link.className).not.toContain('border-brand')
  },
}

export const LongHost: Story = {
  args: {
    integration: makeIntegration({
      id: 'gi3',
      host: 'git.internal.production.acme-platform.example.com',
    }),
  },
}
