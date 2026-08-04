import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import { DeployPill } from './deploy-pill'

const meta = {
  title: 'Features/EditorChrome/DeployPill',
  component: DeployPill,
  tags: ['ai-generated'],
  // The pill positions itself absolutely at the top of the canvas.
  decorators: [
    (Story) => (
      <div className="relative h-24">
        <Story />
      </div>
    ),
  ],
  args: {
    hasResources: true,
    dirtyTotal: 0,
    isStaged: false,
    isActive: false,
    deployBusy: false,
    canWrite: true,
    onDeploy: fn(),
    onViewChanges: fn(),
    canDiscardDraft: false,
  },
} satisfies Meta<typeof DeployPill>

export default meta
type Story = StoryObj<typeof meta>

/** Mid-session dirt: Apply N changes + Details + Deploy + discard menu. */
export const PendingChanges: Story = {
  args: {
    dirtyTotal: 3,
    isActive: true,
    canDiscardDraft: true,
    onDiscardDraft: fn(),
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await expect(canvas.getByText('Apply 3 changes')).toBeInTheDocument()
    await userEvent.click(canvas.getByLabelText('Change actions'))
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Discard draft changes')).toBeInTheDocument()
  },
}

export const Deploying: Story = {
  args: {
    dirtyTotal: 3,
    isActive: true,
    deployBusy: true,
  },
}

/** Draft stack with resources: bare Deploy, no change summary. */
export const Draft: Story = {
  args: {
    isDraft: true,
    onDraftDeploy: fn(),
  },
}

/** Viewer without write access: pill visible, Deploy disabled. */
export const ReadOnly: Story = {
  args: {
    dirtyTotal: 2,
    isActive: true,
    canWrite: false,
  },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('button', { name: /Deploy/ })).toBeDisabled()
  },
}
