import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
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

export const Ready: Story = {
  play: async ({ canvas }) => {
    // Hover/focus stay ink, never brand orange (rubric #3, #8) — mirrors the
    // fix on the sibling DeployStackCard (stack-card.tsx).
    const card = await canvas.findByRole('link', { name: /pr #128/i })
    await expect(card.className).not.toContain('ring-brand')
    await expect(card.className).not.toContain('outline-none')
    await expect(card.className).toMatch(/(?:^|\s)focus-ring(?:-edge|-inset)?(?:\s|$)/)

    const title = await canvas.findByText('PR #128')
    await expect(title.className).not.toContain('text-brand')
  },
}

export const Deploying: Story = {
  args: { env: makeEnv({ status: { phase: 'Deploying' } } as Partial<PreviewStack>) },
}

export const Failed: Story = {
  args: { env: makeEnv({ status: { phase: 'Failed' } } as Partial<PreviewStack>) },
}
