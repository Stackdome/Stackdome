import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Switch } from './switch'

// Zero dedicated story coverage existed for this primitive before this pass
// (flagged in the Task 7 foundation-gate report; picked up here since Task 11
// is the first page pass to touch a form — Add Cluster — that exercises it).
const meta = {
  title: 'Primitives/Switch',
  component: Switch,
  tags: ['ai-generated'],
} satisfies Meta<typeof Switch>

export default meta
type Story = StoryObj<typeof meta>

export const Unchecked: Story = {
  args: { 'aria-label': 'Enable image registry' },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('switch')).toHaveAttribute('data-state', 'unchecked')
  },
}

export const Checked: Story = {
  args: { 'aria-label': 'Enable image registry', defaultChecked: true },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('switch')).toHaveAttribute('data-state', 'checked')
  },
}

export const Disabled: Story = {
  args: { 'aria-label': 'Enable image registry', disabled: true, defaultChecked: true },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('switch')).toBeDisabled()
  },
}

// D13: the thumb's checked/unchecked slide is a state transform, not a hover
// effect — hovering must not move anything further (root or thumb).
export const NoTransformOnHover: Story = {
  args: { 'aria-label': 'Enable image registry' },
  play: async ({ canvas, userEvent }) => {
    const root = canvas.getByRole('switch')
    const before = getComputedStyle(root).transform
    await userEvent.hover(root)
    await expect(getComputedStyle(root).transform).toBe(before)
  },
}

// Focus-visible must render as a solid outline ring off --ring, not the
// removed ring-* utilities — tab to the switch rather than clicking so
// :focus-visible actually engages.
export const KeyboardFocusOutline: Story = {
  args: { 'aria-label': 'Enable image registry' },
  play: async ({ canvas, userEvent }) => {
    const root = canvas.getByRole('switch')
    await userEvent.tab()
    await expect(root).toHaveFocus()
    const style = getComputedStyle(root)
    await expect(style.outlineStyle).not.toBe('none')
    await expect(style.outlineWidth).toBe('2px')
  },
}
