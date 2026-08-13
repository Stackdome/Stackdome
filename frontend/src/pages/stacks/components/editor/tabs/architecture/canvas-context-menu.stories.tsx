import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import { CanvasContextMenu } from './canvas-context-menu'

const meta = {
  title: 'Features/Canvas/CanvasContextMenu',
  component: CanvasContextMenu,
  tags: ['ai-generated'],
  args: {
    onClose: fn(),
    onOpenResource: fn(),
    onAddVolumeToResource: fn(),
    onDeleteResource: fn(),
    onDisconnectVolume: fn(),
    onOpenVolume: fn(),
    onRequestDeleteVolume: fn(),
    onRequestAttach: fn(),
  },
} satisfies Meta<typeof CanvasContextMenu>

export default meta
type Story = StoryObj<typeof meta>

export const ResourceTarget: Story = {
  args: {
    target: { kind: 'resource', resourceIdx: 0, resourceName: 'web', x: 24, y: 24 },
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Open settings')).toBeInTheDocument()
    await expect(body.getByText('Delete service')).toBeInTheDocument()
  },
}

/** Unmounted volume node: offers Attach instead of Disconnect. */
export const VolumeNodeTarget: Story = {
  args: {
    target: { kind: 'volume-node', volumeName: 'web-data', x: 24, y: 24 },
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Attach to service…')).toBeInTheDocument()
  },
}

/** Mounted volume chip: offers Disconnect. */
export const VolumeChipTarget: Story = {
  args: {
    target: { kind: 'volume-chip', volumeName: 'web-data', x: 24, y: 24 },
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Disconnect volume')).toBeInTheDocument()
  },
}
