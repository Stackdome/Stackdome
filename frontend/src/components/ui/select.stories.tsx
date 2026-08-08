import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './select'
import { Button } from './button'
import { Input } from './input'

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

/** §2 — radius is a function of HEIGHT, and the ladder binds selects exactly as
 *  it binds buttons and inputs: 28/6 · 32/8 · 40/12. The trigger used to run a
 *  single 8px corner across every height, so a 28px select in a toolbar sat 2px
 *  rounder than the 28px button beside it. Mirrors Input's RadiusTracksHeight. */
export const RadiusTracksHeight: Story = {
  render: () => (
    <div className="flex flex-col items-start gap-3">
      {(['sm', 'default', 'lg'] as const).map((size) => (
        <Select key={size} defaultValue="Next.js">
          <SelectTrigger size={size} className="w-56" aria-label={size}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {FRAMEWORKS.map((f) => (
              <SelectItem key={f} value={f}>{f}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      ))}
    </div>
  ),
  play: async ({ canvas }) => {
    const expected = { sm: [28, 6], default: [32, 8], lg: [40, 12] } as const
    for (const [size, [height, radius]] of Object.entries(expected)) {
      const el = canvas.getByRole('combobox', { name: size })
      const style = getComputedStyle(el)
      await expect(parseFloat(style.height)).toBe(height)
      await expect(parseFloat(style.borderRadius)).toBe(radius)
    }
  },
}

/** The point of the ladder: at the same height, a button, an input and a select
 *  take the SAME corner. Three unrelated radii in one toolbar row is the kind
 *  of thing the eye catches even when the mind does not (§2). */
export const ControlsAgreeAtTheSameHeight: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button shape="flat" size="sm">Filter</Button>
      <Input size="sm" aria-label="search" placeholder="Search" className="w-40" />
      <Select defaultValue="Next.js">
        <SelectTrigger size="sm" className="w-40" aria-label="framework">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {FRAMEWORKS.map((f) => (
            <SelectItem key={f} value={f}>{f}</SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  ),
  play: async ({ canvas }) => {
    const radii = [
      canvas.getByRole('button', { name: 'Filter' }),
      canvas.getByRole('textbox', { name: 'search' }),
      canvas.getByRole('combobox', { name: 'framework' }),
    ].map((el) => parseFloat(getComputedStyle(el).borderRadius))
    await expect(new Set(radii).size).toBe(1)
    await expect(radii[0]).toBe(6)
  },
}

/**
 * **A select has its own FACE, so hover moves the FILL** — the mirror image of
 * Input, which is a well and moves its line instead.
 *
 * `open` is not a rung of its own: the trigger stays engaged for as long as the
 * menu is out, so it holds the hover fill **and** takes the stronger line. It
 * does not take the press inset — you are not still pushing it.
 */
export const HoverMovesTheFillAndOpenHoldsIt: Story = {
  render: () => (
    <Select defaultValue="Next.js">
      <SelectTrigger className="w-56" aria-label="framework">
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
    const trigger = canvas.getByRole('combobox', { name: 'framework' })
    await expect(trigger.className).toContain('hover:bg-control-hover')
    // It never borrows the wash ladder — that is for faces transparent at rest.
    await expect(trigger.className).not.toContain('wash-hover')

    const rest = getComputedStyle(trigger).backgroundColor
    await userEvent.click(trigger)
    await expect(trigger).toHaveAttribute('data-state', 'open')

    const probe = document.createElement('div')
    probe.style.color = getComputedStyle(document.documentElement).getPropertyValue('--border-strong').trim()
    document.body.appendChild(probe)
    const strongLine = getComputedStyle(probe).color
    probe.remove()

    // The fill and the line are both transitioned, so the first frame after the
    // click still reads the rest values — settle before asserting.
    await waitFor(async () => {
      const open = getComputedStyle(trigger)
      await expect(open.backgroundColor).not.toBe(rest)
      await expect(open.borderColor).toBe(strongLine)
    })

    await userEvent.keyboard('{Escape}')
  },
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
    // The ring is a box-shadow, never an outline — a shadow follows the radius,
    // stacks with the press recess, and can carry a gap when a dark face needs one.
    await expect(style.outlineStyle).toBe('none')
    // Asserted as the CLASS CONTRACT, not the computed pixel: an unlayered
    // `box-shadow` resolves correctly in the app and in a live Storybook — both
    // measured — but comes back as Tailwind's `0 0 #0000` default inside this
    // browser-test harness. Pinning the pixel here would be testing the harness.
    await expect(trigger.className).toMatch(/(?:^|\s)focus-ring(?:-edge|-inset)?(?:\s|$)/)
    await expect(trigger.className).not.toMatch(/(?:^|[\s:])ring-/)
  },
}
