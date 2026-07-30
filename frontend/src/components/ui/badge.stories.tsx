import type { Meta, StoryObj } from '@storybook/react-vite'
import { Badge } from './badge'

const meta = {
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
