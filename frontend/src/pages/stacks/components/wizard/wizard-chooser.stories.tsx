import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { WizardChooser } from './wizard-chooser'

const meta = {
  title: 'Features/Wizard/WizardChooser',
  component: WizardChooser,
  tags: ['ai-generated'],
} satisfies Meta<typeof WizardChooser>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    onPickBlocks: fn(),
    onPickTemplate: fn(),
    onPickCompose: fn(),
    onPickBlank: fn(),
    onPickGit: fn(),
  },
}
