import type { Meta, StoryObj } from '@storybook/react-vite'
import { StatusPill } from './status-pill'

const meta = {
  component: StatusPill,
  tags: ['ai-generated'],
} satisfies Meta<typeof StatusPill>

export default meta
type Story = StoryObj<typeof meta>

export const Ready: Story = { args: { variant: 'ready', children: 'Ready' } }
export const Pending: Story = { args: { variant: 'pending', children: 'Provisioning' } }
export const Error: Story = { args: { variant: 'error', children: 'Failed' } }
export const Info: Story = { args: { variant: 'info', children: 'Deploying' } }
export const Neutral: Story = { args: { variant: 'neutral', children: 'Idle' } }
export const WithoutDot: Story = { args: { variant: 'ready', withDot: false, children: 'Ready' } }
export const PulseSuppressed: Story = {
  args: { variant: 'ready', pulse: false, children: 'Backup · 2d ago' },
}
