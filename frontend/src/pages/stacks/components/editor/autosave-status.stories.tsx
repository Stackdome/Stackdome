import type { Meta, StoryObj } from '@storybook/react-vite'
import { SYNC_STATUS } from '@/pages/stacks/lib/draft-sync/constants'
import { AutosaveStatus } from './autosave-status'

const meta = {
  title: 'Features/EditorChrome/AutosaveStatus',
  component: AutosaveStatus,
  tags: ['ai-generated'],
} satisfies Meta<typeof AutosaveStatus>

export default meta
type Story = StoryObj<typeof meta>

export const Saving: Story = {
  args: { status: SYNC_STATUS.saving },
}

export const Saved: Story = {
  args: { status: SYNC_STATUS.saved },
}

export const Error: Story = {
  args: { status: SYNC_STATUS.error },
}
