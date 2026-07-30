import type { Meta, StoryObj } from '@storybook/react-vite'
import { Button } from '@/components/ui/button'
import { PageHeader } from './page-header'
import { StatusPill } from './status-pill'

const meta = {
  component: PageHeader,
  tags: ['ai-generated'],
  args: { title: 'Stacks' },
} satisfies Meta<typeof PageHeader>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
export const Full: Story = {
  args: {
    eyebrow: 'Infrastructure',
    title: 'orders-api',
    status: <StatusPill variant="ready">Ready</StatusPill>,
    subtitle: 'prod-cluster · eu-west-1 · last deployed 4m ago',
    actions: (
      <>
        <Button variant="outline">Settings</Button>
        <Button>Deploy</Button>
      </>
    ),
  },
}
export const ActionsCentered: Story = {
  args: {
    title: 'Clusters',
    subtitle: '3 connected',
    actionsAlign: 'center',
    actions: <Button>Connect cluster</Button>,
  },
}
