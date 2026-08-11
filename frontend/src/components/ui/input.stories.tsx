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
export const Filled: Story = { args: { defaultValue: 'my-first-stack' } }
export const Disabled: Story = { args: { disabled: true, defaultValue: 'my-first-stack' } }
export const Invalid: Story = {
  args: { 'aria-invalid': true, defaultValue: 'not a valid name!' },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('textbox')).toHaveAttribute('aria-invalid', 'true')
  },
}
export const LongValueOverflow: Story = {
  args: {
    defaultValue:
      'this-is-an-unreasonably-long-stack-name-that-should-not-blow-out-the-pill-shaped-field-layout',
  },
  render: (args) => (
    <div className="w-64">
      <Input {...args} />
    </div>
  ),
}
export const Password: Story = { args: { type: 'password', defaultValue: 'hunter2' } }

// Focus-visible must render as a solid outline ring off --ring, not the
// removed ring-* utilities — tab to the input rather than clicking so
// :focus-visible actually engages. Mirrors Button's KeyboardFocusOutline.
export const KeyboardFocusOutline: Story = {
  play: async ({ canvas, userEvent }) => {
    const input = canvas.getByRole('textbox')
    await userEvent.tab()
    await expect(input).toHaveFocus()
    const style = getComputedStyle(input)
    await expect(style.outlineStyle).not.toBe('none')
    await expect(style.outlineWidth).toBe('2px')
  },
}
