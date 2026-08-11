import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { SecretList } from './secret-list'
import type { Secret } from '../types'

function makeSecret(overrides: Partial<Secret> = {}): Secret {
  return {
    id: 's1',
    name: 'stripe-api-key',
    description: 'Live Stripe secret key',
    type: 'Token',
    data: [],
    created_at: '2026-06-01T10:00:00Z',
    ...overrides,
  }
}

const meta = {
  title: 'Features/Secrets/SecretList',
  component: SecretList,
  tags: ['ai-generated'],
  args: { onEdit: fn(), onDelete: fn() },
  decorators: [
    (Story) => (
      <div className="max-w-[820px] rounded-md border">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SecretList>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    secrets: [
      makeSecret(),
      makeSecret({ id: 's2', name: 'db-password', type: 'UsernamePassword', description: '' }),
      makeSecret({ id: 's3', name: 'deploy-key', type: 'SSHKey' }),
    ],
  },
  play: async ({ canvas }) => {
    const trigger = canvas.getAllByRole('button', { name: 'Secret actions' })[0]
    // One control height (rubric #9): the row menu trigger reads from the
    // Button `icon` size, never a hand-set h-8/w-8 override.
    await expect(trigger.className).toContain('size-10')
    await expect(trigger.className).not.toMatch(/\bh-8\b/)
    await expect(trigger.className).toContain('focus-visible:outline-2')
  },
}

export const ReadOnly: Story = {
  args: {
    secrets: [makeSecret()],
    canWrite: () => false,
  },
}

export const LongNameAndDescription: Story = {
  args: {
    secrets: [
      makeSecret({
        id: 's4',
        name: 'production-payment-gateway-webhook-signing-secret-rotated',
        description:
          'Used to verify inbound webhook signatures from the payment processor across every environment',
      }),
    ],
  },
}

export const Empty: Story = {
  args: { secrets: [] },
}
