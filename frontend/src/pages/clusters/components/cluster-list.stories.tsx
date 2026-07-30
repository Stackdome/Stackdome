import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
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
}

export const Empty: Story = {
  args: { clusters: [] },
}
