import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { AlertBanner } from './alert-banner'

const meta = {
  title: 'Branded/AlertBanner',
  component: AlertBanner,
  tags: ['ai-generated'],
  args: { children: 'Deployment failed: image pull backoff on web-1.' },
} satisfies Meta<typeof AlertBanner>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
export const WithAction: Story = {
  args: { action: { label: 'Retry', onClick: fn() } },
  play: async ({ canvas, userEvent, args }) => {
    await userEvent.click(canvas.getByRole('button', { name: /retry/i }))
    await expect(args.action!.onClick).toHaveBeenCalled()
  },
}
export const ActionDisabled: Story = {
  args: { action: { label: 'Retry', onClick: fn(), disabled: true } },
}
