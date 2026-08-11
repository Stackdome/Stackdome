import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { makeCluster } from '../../../../.storybook/fixtures'
import { ClusterList } from './cluster-list'

const meta = {
  title: 'Features/Clusters/ClusterList',
  component: ClusterList,
  tags: ['ai-generated'],
  args: { onOpen: fn() },
  decorators: [
    (Story) => (
      <div className="max-w-[640px] rounded-md border">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ClusterList>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    clusters: [
      makeCluster(),
      makeCluster({ id: 'c2', name: 'staging-eu-west' }),
      makeCluster({ id: 'c3', name: 'edge-ap-south' }),
    ],
  },
  play: async ({ canvas }) => {
    const rows = canvas.getAllByRole('button')
    for (const row of rows) {
      // Focus is a stylesheet outline off --ring, never the removed ring-*
      // utilities, and nothing on the row moves on hover/press (rubric #6/#10).
      await expect(row.className).toContain('focus-visible:outline-2')
      await expect(row.className).not.toContain('ring-')
      await expect(row.className).not.toContain('translate')
      await expect(row.className).not.toContain('scale')
    }
    // The leading glyph and chevron are decorative chrome, not focusable
    // controls in their own right (rubric #14) — the row is the one control.
    const icon = canvas.getAllByRole('button')[0].querySelector('svg')
    await expect(icon?.getAttribute('aria-hidden')).not.toBeNull()
  },
}

export const Empty: Story = {
  args: { clusters: [] },
}

export const LongNameAndId: Story = {
  args: {
    clusters: [
      makeCluster({
        id: 'c4',
        name: 'production-us-east-1-primary-multi-az-autoscaling-cluster',
      }),
    ],
  },
}
