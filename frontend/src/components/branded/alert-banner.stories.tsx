import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { AlertBanner } from './alert-banner'

const meta = {
  title: 'Branded/AlertBanner',
  component: AlertBanner,
  tags: ['ai-generated'],
  args: { children: 'Deployment failed: image pull backoff on web-1.' },
  decorators: [
    (Story) => (
      <div className="bg-card w-[620px] max-w-full p-4">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof AlertBanner>

export default meta
type Story = StoryObj<typeof meta>

/** `danger` is the default, so every caller that predates tones is unchanged. */
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

/** `blocking` — it has not failed yet, but it will unless you deal with this. */
export const Blocking: Story = {
  args: {
    tone: 'blocking',
    children: 'A database password is written into the file in plain text. Move it to a secret before deploying.',
    action: { label: 'Move to a secret', onClick: fn() },
  },
}

/** `info` — we changed something about your input, and you should know. */
export const Info: Story = {
  args: {
    tone: 'info',
    children: 'Host port bindings were ignored. Stackdome assigns the route itself.',
    action: { label: 'Dismiss', onClick: fn() },
  },
}

/** All three together — proof that severity reads off the hue alone, and that
 *  the message stays ink in every one of them (§7). */
export const EveryTone: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      <AlertBanner>Deployment failed: image pull backoff on web-1.</AlertBanner>
      <AlertBanner tone="blocking">
        A database password is written into the file in plain text. Move it to a secret before deploying.
      </AlertBanner>
      <AlertBanner tone="info">Host port bindings were ignored. Stackdome assigns the route itself.</AlertBanner>
    </div>
  ),
}

/** Long copy wraps and the action holds its place on the right. */
export const LongMessageWithAction: Story = {
  args: {
    tone: 'blocking',
    children:
      'A database password is written into the compose file in plain text, and two more services reference it by the same literal. Move all three to a secret before deploying, or the value ends up in the build log.',
    action: { label: 'Move to a secret', onClick: fn() },
  },
}
