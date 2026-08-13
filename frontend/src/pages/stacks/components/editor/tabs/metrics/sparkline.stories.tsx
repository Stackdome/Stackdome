import type { Meta, StoryObj } from '@storybook/react-vite'
import { Sparkline } from './metrics-tab'

const meta = {
  title: 'Features/Metrics/Sparkline',
  component: Sparkline,
  tags: ['ai-generated'],
} satisfies Meta<typeof Sparkline>

export default meta
type Story = StoryObj<typeof meta>

export const Trend: Story = {
  args: {
    data: [4, 6, 5, 8, 9, 12, 11, 14, 16, 18, 17, 21, 24, 23, 27, 30],
    className: 'bg-brand',
  },
}

export const Flat: Story = {
  args: {
    data: Array(16).fill(10),
    className: 'bg-brand',
  },
}

// One sample so far: 15 placeholder stubs, one full-height bar at the end.
export const SinglePoint: Story = {
  args: {
    data: [42],
    className: 'bg-fg-2',
  },
}
