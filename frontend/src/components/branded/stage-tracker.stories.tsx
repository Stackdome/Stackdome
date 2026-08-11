import type { Meta, StoryObj } from '@storybook/react-vite'
import { StageTracker } from './stage-tracker'

const meta = {
  title: 'Features/Deployments/StageTracker',
  component: StageTracker,
  tags: ['ai-generated'],
} satisfies Meta<typeof StageTracker>

export default meta
type Story = StoryObj<typeof meta>

export const AllStages: Story = {
  args: { stages: { build: 'done', deploy: 'done', ready: 'done' } },
}

export const FailedMidway: Story = {
  args: { stages: { build: 'done', deploy: 'failed', ready: 'todo' } },
}

// Image-only stack: no build step — solid muted "skipped" fill, not a hollow todo ring.
export const BuildSkipped: Story = {
  args: { stages: { build: 'skipped', deploy: 'active', ready: 'todo' } },
}
