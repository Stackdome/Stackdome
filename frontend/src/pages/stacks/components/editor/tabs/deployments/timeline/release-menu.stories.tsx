import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import { makeRelease } from '../../../../../../../../.storybook/fixtures'
import { ReleaseState } from '../release-states'
import { ReleaseMenu } from './release-menu'

const meta = {
  title: 'Features/Deployments/ReleaseMenu',
  component: ReleaseMenu,
  tags: ['ai-generated'],
  args: { onRollback: fn(), onCancel: fn(), onCopyId: fn() },
} satisfies Meta<typeof ReleaseMenu>

export default meta
type Story = StoryObj<typeof meta>

// Released → rollback offered; content renders in a Radix portal on document.body.
export const Default: Story = {
  args: { release: makeRelease() },
  play: async ({ canvas, canvasElement, userEvent, args }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Release actions' }))
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Rollback to this')).toBeInTheDocument()
    await expect(body.getByText('Copy release ID')).toBeInTheDocument()
    await userEvent.click(body.getByText('Rollback to this'))
    await expect(args.onRollback).toHaveBeenCalledWith('rel-12')
  },
}

// Only a Pending release can still be cancelled; rollback is absent.
export const Pending: Story = {
  args: { release: makeRelease({ id: 'rel-13', sequence: 13, state: ReleaseState.Pending, completed_at: undefined }) },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Release actions' }))
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Cancel release')).toBeInTheDocument()
    await expect(body.queryByText('Rollback to this')).toBeNull()
  },
}
