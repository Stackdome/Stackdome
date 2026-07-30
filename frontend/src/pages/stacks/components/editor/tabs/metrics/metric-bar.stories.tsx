import type { Meta, StoryObj } from '@storybook/react-vite'
import { MetricBar } from './metrics-tab'

const meta = {
  title: 'Features/Metrics/MetricBar',
  component: MetricBar,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[280px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof MetricBar>

export default meta
type Story = StoryObj<typeof meta>

export const Low: Story = {
  args: { label: 'CPU', value: '45m', pct: 12, fill: 'bg-brand' },
}

export const High: Story = {
  args: { label: 'Memory', value: '1.4 Gi', pct: 85, fill: 'bg-fg-2' },
}

// pct > 100 clamps to a full bar rather than overflowing the track.
export const Over: Story = {
  args: { label: 'CPU', value: '2.3 cores', pct: 130, fill: 'bg-brand' },
}
