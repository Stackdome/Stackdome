import type { Meta, StoryObj } from '@storybook/react-vite'
import { Database } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { EmptyState } from './empty-state'

const meta = {
  component: EmptyState,
  tags: ['ai-generated'],
  args: { title: 'No stacks yet' },
} satisfies Meta<typeof EmptyState>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
export const WithEverything: Story = {
  args: {
    icon: <Database className="h-8 w-8" />,
    title: 'No addons provisioned',
    description: 'Provision a managed Postgres to attach it to your stacks.',
    action: <Button size="sm">Create Postgres</Button>,
  },
}
export const Solid: Story = { args: { dashed: false, title: 'Nothing to show' } }
