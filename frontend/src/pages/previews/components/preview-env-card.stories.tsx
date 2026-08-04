import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import type { PreviewStack } from '@/api/preview-envs'
import { PreviewEnvCard } from './preview-env-card'

const makeEnv = (overrides: Partial<PreviewStack> = {}): PreviewStack =>
  ({
    id: 'pe1',
    pr_number: 128,
    stack_id: 's9',
    status: {
      phase: 'Ready',
      outputs: { urls: [{ resource: 'web', url: 'https://pr-128.preview.example.com' }] },
    },
    updated_at: '2026-07-30T12:00:00Z',
    ...overrides,
  }) as PreviewStack

const meta = {
  title: 'Features/StackCard/PreviewEnvCard',
  component: PreviewEnvCard,
  tags: ['ai-generated'],
  args: { env: makeEnv(), onSync: fn(), onDelete: fn() },
  decorators: [
    (Story) => (
      <div className="max-w-[380px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof PreviewEnvCard>

export default meta
type Story = StoryObj<typeof meta>

export const Ready: Story = {}

export const Deploying: Story = {
  args: { env: makeEnv({ status: { phase: 'Deploying' } } as Partial<PreviewStack>) },
}

export const Failed: Story = {
  args: { env: makeEnv({ status: { phase: 'Failed' } } as Partial<PreviewStack>) },
}
