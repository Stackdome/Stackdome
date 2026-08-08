import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Checkbox } from './checkbox'

const meta = {
  title: 'Primitives/Checkbox',
  component: Checkbox,
} satisfies Meta<typeof Checkbox>

export default meta
type Story = StoryObj<typeof meta>

export const Unchecked: Story = { args: { 'aria-label': 'Acknowledge' } }
export const Checked: Story = { args: { 'aria-label': 'Acknowledge', defaultChecked: true } }
export const Indeterminate: Story = { args: { 'aria-label': 'Acknowledge', indeterminate: true } }
export const Disabled: Story = { args: { 'aria-label': 'Acknowledge', disabled: true } }
export const DisabledChecked: Story = {
  args: { 'aria-label': 'Acknowledge', disabled: true, defaultChecked: true },
}

export const WithLabel: Story = {
  args: {},
  render: () => (
    <label className="flex max-w-sm items-start gap-2.5 text-body text-fg-2">
      <Checkbox className="mt-0.5" />
      <span>I understand that services using this key will start failing immediately.</span>
    </label>
  ),
  play: async ({ canvas, userEvent }) => {
    const box = canvas.getByRole('checkbox')
    await expect(box).not.toBeChecked()
    // The label is the hit target too — a 16px box is not one on its own.
    await userEvent.click(canvas.getByText(/start failing immediately/))
    await expect(box).toBeChecked()
  },
}

/** Controlled use, which is how the §6a acknowledge gate drives it. */
export const Controlled: Story = {
  args: {},
  render: () => {
    const [on, setOn] = useState(false)
    return (
      <div className="flex items-center gap-2">
        <Checkbox aria-label="Acknowledge" checked={on} onCheckedChange={setOn} />
        <span data-testid="state" className="text-body">{String(on)}</span>
      </div>
    )
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('checkbox'))
    await expect(canvas.getByTestId('state')).toHaveTextContent('true')
    await userEvent.click(canvas.getByRole('checkbox'))
    await expect(canvas.getByTestId('state')).toHaveTextContent('false')
  },
}

/** It is a real `<input type="checkbox">`, so the keyboard works without any
 *  help — space toggles, and Tab reaches it. That is the reason the component
 *  is a styled box over a native control rather than a div pretending. */
export const KeyboardOperable: Story = {
  args: { 'aria-label': 'Acknowledge' },
  play: async ({ canvas, userEvent }) => {
    const box = canvas.getByRole('checkbox')
    await userEvent.tab()
    await expect(box).toHaveFocus()
    await userEvent.keyboard(' ')
    await expect(box).toBeChecked()
    await userEvent.keyboard(' ')
    await expect(box).not.toBeChecked()
  },
}
