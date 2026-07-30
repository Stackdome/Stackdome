import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Input } from './input'

const meta = {
  title: 'Primitives/Input',
  component: Input,
  tags: ['ai-generated'],
  args: { placeholder: 'my-stack-name' },
} satisfies Meta<typeof Input>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
export const Disabled: Story = { args: { disabled: true } }
export const Invalid: Story = {
  args: { 'aria-invalid': true, defaultValue: 'not a valid name!' },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('textbox')).toHaveAttribute('aria-invalid', 'true')
  },
}
export const Password: Story = { args: { type: 'password', defaultValue: 'hunter2' } }
