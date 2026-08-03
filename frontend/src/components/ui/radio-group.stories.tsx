import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { RadioGroup, RadioGroupItem } from './radio-group'

// Zero dedicated story coverage existed for this primitive before this pass
// (flagged in Tasks 7/11 as deferred; picked up here since Task 12 — Addons —
// is the first page pass whose form uses it, for the plan-select table on
// the create/edit page). Unchecked border: settled on `border-border-strong`
// (already the primitive's default) over the plain `--border` hairline —
// against `bg-card`/table-row surfaces the 11%-alpha hairline nearly
// vanishes at the 16px dot size, so the primitive keeps the stronger 18%
// tier rather than the page patching it per call site.
const meta = {
  title: 'Primitives/RadioGroup',
  component: RadioGroup,
  tags: ['ai-generated'],
} satisfies Meta<typeof RadioGroup>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <RadioGroup defaultValue="standard" aria-label="Plan">
      <label className="flex items-center gap-2 text-sm">
        <RadioGroupItem value="standard" id="rg-standard" />
        Standard
      </label>
      <label className="flex items-center gap-2 text-sm">
        <RadioGroupItem value="custom" id="rg-custom" />
        Custom
      </label>
    </RadioGroup>
  ),
  play: async ({ canvas }) => {
    const options = canvas.getAllByRole('radio')
    await expect(options[0]).toHaveAttribute('data-state', 'checked')
    await expect(options[1]).toHaveAttribute('data-state', 'unchecked')
  },
}

export const Disabled: Story = {
  render: () => (
    <RadioGroup defaultValue="standard" disabled aria-label="Plan">
      <label className="flex items-center gap-2 text-sm">
        <RadioGroupItem value="standard" id="rg-standard-disabled" />
        Standard
      </label>
    </RadioGroup>
  ),
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('radio')).toBeDisabled()
  },
}

// The unchecked dot must stay visible against a flat card/table surface —
// this is the decision recorded above, pinned as a regression check.
export const UncheckedBorderVisible: Story = {
  render: () => (
    <RadioGroup aria-label="Plan">
      <RadioGroupItem value="only" id="rg-visible" />
    </RadioGroup>
  ),
  play: async ({ canvas }) => {
    const item = canvas.getByRole('radio')
    const strong = getComputedStyle(document.documentElement).getPropertyValue('--border-strong').trim()
    const probe = document.createElement('div')
    probe.style.borderColor = strong
    document.body.appendChild(probe)
    const expected = getComputedStyle(probe).borderColor
    probe.remove()
    await expect(getComputedStyle(item).borderColor).toBe(expected)
  },
}

// D13: selecting/deselecting is a state change, not a hover effect —
// hovering an unselected radio must not move or scale it.
export const NoTransformOnHover: Story = {
  render: () => (
    <RadioGroup aria-label="Plan">
      <RadioGroupItem value="only" id="rg-hover" />
    </RadioGroup>
  ),
  play: async ({ canvas, userEvent }) => {
    const item = canvas.getByRole('radio')
    const before = getComputedStyle(item).transform
    await userEvent.hover(item)
    await expect(getComputedStyle(item).transform).toBe(before)
  },
}

// Focus-visible must render as a solid outline off --ring, not a box-shadow
// ring — tab to the control rather than clicking so :focus-visible engages.
export const KeyboardFocusOutline: Story = {
  render: () => (
    <RadioGroup aria-label="Plan">
      <RadioGroupItem value="only" id="rg-focus" />
    </RadioGroup>
  ),
  play: async ({ canvas, userEvent }) => {
    const item = canvas.getByRole('radio')
    await userEvent.tab()
    await expect(item).toHaveFocus()
    const style = getComputedStyle(item)
    await expect(style.outlineStyle).not.toBe('none')
    await expect(style.outlineWidth).toBe('2px')
  },
}
