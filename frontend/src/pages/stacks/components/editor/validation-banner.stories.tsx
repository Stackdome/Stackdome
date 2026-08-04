import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { ValidationBanner } from './validation-banner'

const meta = {
  title: 'Features/EditorChrome/ValidationBanner',
  component: ValidationBanner,
  tags: ['ai-generated'],
  args: {
    onJump: fn(),
    onDismiss: fn(),
  },
} satisfies Meta<typeof ValidationBanner>

export default meta
type Story = StoryObj<typeof meta>

/** Mix of jumpable resource errors and a stack-level one (no chevron). */
export const Errors: Story = {
  args: {
    items: [
      { label: 'api', message: 'image is required when no build source is set', resourceIndex: 0 },
      { label: 'worker', message: 'port 8080 is already in use by api', resourceIndex: 1, tab: 'configuration' },
      { label: 'Stack settings', message: 'a stack with this name already exists' },
    ],
  },
}

export const SingleError: Story = {
  args: {
    items: [{ label: 'api', message: 'image is required when no build source is set', resourceIndex: 0 }],
  },
}
