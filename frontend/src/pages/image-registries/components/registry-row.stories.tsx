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
    // Row-menu trigger reads the shared icon-button size — no hand-set h-8/w-8
    // override. The icon default is 32px since the control-height ladder landed.
    await expect(trigger.className).toContain('size-8')
    await expect(trigger.className).not.toMatch(/\bh-8\b/)
    await expect(trigger.className).toContain('focus-ring')

    // A row menu is a working control, so it is `flat`, and §2 makes its radius
    // a function of its height: 32px takes 8px. That 8 comes from the variant
    // ladder — this used to assert `rounded-md` was ABSENT, back when a radius
    // in the class list could only have been hand-set at the call site.
    const style = getComputedStyle(trigger)
    await expect(parseFloat(style.height)).toBe(32)
    await expect(parseFloat(style.borderRadius)).toBe(8)
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
