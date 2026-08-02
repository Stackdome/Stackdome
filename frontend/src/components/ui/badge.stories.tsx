import type { Meta, StoryObj } from '@storybook/react-vite'
import { Badge } from './badge'

const meta = {
  title: 'Primitives/Badge',
  component: Badge,
  tags: ['ai-generated'],
  args: { children: 'v1.4.2' },
} satisfies Meta<typeof Badge>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
export const Secondary: Story = { args: { variant: 'secondary' } }
export const Destructive: Story = { args: { variant: 'destructive', children: 'Failed' } }
export const Outline: Story = { args: { variant: 'outline' } }
export const Success: Story = { args: { variant: 'success', children: 'Ready' } }
export const Warning: Story = { args: { variant: 'warning', children: 'Degraded' } }
export const Info: Story = { args: { variant: 'info', children: 'Provisioning' } }
