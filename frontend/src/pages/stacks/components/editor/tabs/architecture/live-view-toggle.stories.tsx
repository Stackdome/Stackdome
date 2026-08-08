import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { LiveViewToggle } from './live-view-toggle'

const meta = {
  title: 'Features/Canvas/LiveViewToggle',
  component: LiveViewToggle,
  tags: ['ai-generated'],
  args: { onModeChange: fn() },
} satisfies Meta<typeof LiveViewToggle>

export default meta
type Story = StoryObj<typeof meta>

export const Draft: Story = {
  args: { mode: 'draft', draftDirty: false },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('button', { name: /draft/i })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  },
}

export const Live: Story = {
  args: { mode: 'live', draftDirty: false },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('button', { name: /live/i })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  },
}

// Undeployed edits mark the Draft segment so switching to Live doesn't read
// as "nothing pending".
export const DraftDirty: Story = {
  args: { mode: 'live', draftDirty: true },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector('[data-testid="draft-dirty-dot"]')).toBeTruthy()
  },
}
