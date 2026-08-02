import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './select'

const FRAMEWORKS = ['Next.js', 'Remix', 'Astro', 'SvelteKit', 'Nuxt']

const LONG_LIST = Array.from({ length: 30 }, (_, i) => `cluster-region-${i + 1}`)

const meta = {
  title: 'Primitives/Select',
  component: Select,
  tags: ['ai-generated'],
} satisfies Meta<typeof Select>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Select defaultValue="Next.js">
      <SelectTrigger className="w-56">
        <SelectValue placeholder="Select a framework" />
      </SelectTrigger>
      <SelectContent>
        {FRAMEWORKS.map((f) => (
          <SelectItem key={f} value={f}>{f}</SelectItem>
        ))}
      </SelectContent>
    </Select>
  ),
}

export const Disabled: Story = {
  render: () => (
    <Select disabled defaultValue="Next.js">
      <SelectTrigger className="w-56">
        <SelectValue placeholder="Select a framework" />
      </SelectTrigger>
      <SelectContent>
        {FRAMEWORKS.map((f) => (
          <SelectItem key={f} value={f}>{f}</SelectItem>
        ))}
      </SelectContent>
    </Select>
  ),
}

export const LongList: Story = {
  render: () => (
    <Select defaultValue={LONG_LIST[0]}>
      <SelectTrigger className="w-56">
        <SelectValue placeholder="Select a region" />
      </SelectTrigger>
      <SelectContent>
        {LONG_LIST.map((region) => (
          <SelectItem key={region} value={region}>{region}</SelectItem>
        ))}
      </SelectContent>
    </Select>
  ),
}

// aria-invalid drives border-danger via the same token the Button CssCheck
// story reads live — keeps this passing as the palette evolves.
export const Invalid: Story = {
  render: () => (
    <Select defaultValue="Next.js">
      <SelectTrigger className="w-56" aria-invalid>
        <SelectValue placeholder="Select a framework" />
      </SelectTrigger>
      <SelectContent>
        {FRAMEWORKS.map((f) => (
          <SelectItem key={f} value={f}>{f}</SelectItem>
        ))}
      </SelectContent>
    </Select>
  ),
  play: async ({ canvas }) => {
    const trigger = canvas.getByRole('combobox')
    await expect(trigger).toHaveAttribute('aria-invalid', 'true')
    const probe = document.createElement('div')
    probe.style.color = getComputedStyle(document.documentElement).getPropertyValue('--danger').trim()
    document.body.appendChild(probe)
    const expected = getComputedStyle(probe).color
    probe.remove()
    await expect(getComputedStyle(trigger).borderColor).toBe(expected)
  },
}

// Focus-visible must render as a solid outline ring off --ring, not the
// removed ring-* utilities — tab to the trigger rather than clicking so
// :focus-visible actually engages. Mirrors Button's KeyboardFocusOutline.
export const KeyboardFocusOutline: Story = {
  render: () => (
    <Select defaultValue="Next.js">
      <SelectTrigger className="w-56">
        <SelectValue placeholder="Select a framework" />
      </SelectTrigger>
      <SelectContent>
        {FRAMEWORKS.map((f) => (
          <SelectItem key={f} value={f}>{f}</SelectItem>
        ))}
      </SelectContent>
    </Select>
  ),
  play: async ({ canvas, userEvent }) => {
    const trigger = canvas.getByRole('combobox')
    await userEvent.tab()
    await expect(trigger).toHaveFocus()
    const style = getComputedStyle(trigger)
    await expect(style.outlineStyle).not.toBe('none')
    await expect(style.outlineWidth).toBe('2px')
  },
}
