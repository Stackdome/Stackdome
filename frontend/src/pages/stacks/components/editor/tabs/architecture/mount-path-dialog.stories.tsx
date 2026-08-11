import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import { MountPathDialog } from './mount-path-dialog'

const meta = {
  title: 'Features/Canvas/MountPathDialog',
  component: MountPathDialog,
  tags: ['ai-generated'],
  args: {
    resources: [{ name: 'web' }, { name: 'api' }],
    onCancel: fn(),
    onAttach: fn(),
  },
} satisfies Meta<typeof MountPathDialog>

export default meta
type Story = StoryObj<typeof meta>

/** Opened from the menu: service picker + mount path. */
export const Open: Story = {
  args: {
    volumeName: 'web-data',
    resourceIdx: null,
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByRole('dialog')).toBeInTheDocument()
    await expect(body.getByLabelText(/mount path/i)).toBeInTheDocument()
    await expect(body.getByText('Select service')).toBeInTheDocument()
  },
}

/** Opened by drag-drop onto a service: target fixed, no picker. */
export const FixedTarget: Story = {
  args: {
    volumeName: 'web-data',
    resourceIdx: 0,
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByRole('dialog')).toBeInTheDocument()
    await expect(body.queryByText('Select service')).toBeNull()
  },
}
