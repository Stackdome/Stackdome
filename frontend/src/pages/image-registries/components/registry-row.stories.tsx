import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { RegistryRow } from './registry-row'
import type { RegistryCredential } from '@/api/registry-credentials'

function makeCredential(overrides: Partial<RegistryCredential> = {}): RegistryCredential {
  return {
    id: 'r1',
    host: 'index.docker.io',
    purpose: 'both',
    username: 'acme-bot',
    created_at: '2026-06-01T10:00:00Z',
    ...overrides,
  }
}

const meta = {
  title: 'Features/ImageRegistries/RegistryRow',
  component: RegistryRow,
  tags: ['ai-generated'],
  args: { onVerify: fn(), onUpdateCredentials: fn(), onRemove: fn() },
  decorators: [
    (Story) => (
      <div className="max-w-[820px] divide-y divide-border rounded-md border">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof RegistryRow>

export default meta
type Story = StoryObj<typeof meta>

export const DockerHub: Story = {
  args: { credential: makeCredential() },
  play: async ({ canvas }) => {
    const trigger = canvas.getByRole('button', { name: 'Open row menu' })
    // Row-menu trigger reads the shared icon-button size — no hand-set
    // h-8/w-8 override, no arbitrary radius swap (rubric #9).
    await expect(trigger.className).toContain('size-10')
    await expect(trigger.className).not.toMatch(/\bh-8\b/)
    await expect(trigger.className).not.toContain('rounded-md')
    await expect(trigger.className).toContain('focus-visible:outline-2')
  },
}

export const Ghcr: Story = {
  args: { credential: makeCredential({ id: 'r2', host: 'ghcr.io', purpose: 'pull' }) },
}

export const CustomHostLongName: Story = {
  args: {
    credential: makeCredential({
      id: 'r3',
      host: 'registry.internal.production.acme-platform.example.com',
      purpose: 'push',
      username: 'ci-deploy-service-account-bot',
    }),
  },
}
