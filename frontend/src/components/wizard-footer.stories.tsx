import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { WizardFooter } from './wizard-footer'

const meta = {
  title: 'Features/Wizard/WizardFooter',
  component: WizardFooter,
  tags: ['ai-generated'],
  args: { onBack: fn(), onContinue: fn() },
  decorators: [
    (Story) => (
      <div className="max-w-[720px] border">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof WizardFooter>

export default meta
type Story = StoryObj<typeof meta>

export const FirstStep: Story = {}

export const LastStep: Story = {
  args: {
    continueLabel: 'Create stack',
    hint: 'Open in the canvas editor',
  },
}

export const Busy: Story = {
  args: { continueLabel: 'Create stack', loading: true },
}

export const Disabled: Story = {
  args: { continueDisabled: true, hint: 'Pick at least one block' },
}
