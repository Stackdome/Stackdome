import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { LayoutGrid, List, Rows3 } from 'lucide-react'
import { SegmentedControl, type SegmentedControlOption } from './segmented-control'

const VIEWS: SegmentedControlOption<'list' | 'cards'>[] = [
  { value: 'list', label: 'List', icon: <List /> },
  { value: 'cards', label: 'Cards', icon: <LayoutGrid /> },
]

const DENSITY: SegmentedControlOption<'compact' | 'cosy' | 'roomy'>[] = [
  { value: 'compact', label: 'Compact' },
  { value: 'cosy', label: 'Cosy' },
  { value: 'roomy', label: 'Roomy' },
]

/** Stateful host — the control is controlled, so a story that never updates
 *  `value` would prove the click handler fires and nothing else. */
function Harness<T extends string>({
  options,
  initial,
  ...rest
}: {
  options: SegmentedControlOption<T>[]
  initial: T
  size?: 'sm' | 'default'
  disabled?: boolean
  'aria-label'?: string
}) {
  const [value, setValue] = useState<T>(initial)
  return <SegmentedControl options={options} value={value} onValueChange={setValue} {...rest} />
}

const meta = {
  title: 'Primitives/SegmentedControl',
  component: SegmentedControl,
  // The control is controlled, so every story renders through Harness. These
  // args exist only to satisfy the required props on the type.
  args: { options: VIEWS, value: 'list', onValueChange: () => {} },
  render: () => <Harness options={VIEWS} initial="list" aria-label="View" />,
} satisfies Meta<typeof SegmentedControl>

export default meta
type Story = StoryObj<typeof meta>

export const IconOnly: Story = {
  render: () => <Harness options={VIEWS} initial="list" aria-label="View" />,
}

export const Labelled: Story = {
  render: () => <Harness options={DENSITY} initial="cosy" aria-label="Density" />,
}

export const IconAndLabel: Story = {
  render: () => (
    <Harness
      options={VIEWS.map((o) => ({ ...o, showLabel: true }))}
      initial="list"
      aria-label="View"
    />
  ),
}

export const ThreeSegments: Story = {
  render: () => (
    <Harness
      options={[
        { value: 'list', label: 'List', icon: <List /> },
        { value: 'rows', label: 'Rows', icon: <Rows3 /> },
        { value: 'cards', label: 'Cards', icon: <LayoutGrid /> },
      ]}
      initial="rows"
      aria-label="View"
    />
  ),
}

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Harness options={VIEWS} initial="list" size="sm" aria-label="View small" />
      <Harness options={VIEWS} initial="list" size="default" aria-label="View default" />
    </div>
  ),
}

export const Disabled: Story = {
  render: () => <Harness options={VIEWS} initial="list" disabled aria-label="View" />,
  play: async ({ canvas }) => {
    for (const name of ['List', 'Cards']) {
      await expect(canvas.getByRole('radio', { name })).toBeDisabled()
    }
  },
}

export const OneSegmentDisabled: Story = {
  render: () => (
    <Harness
      options={[
        { value: 'list', label: 'List', icon: <List /> },
        { value: 'cards', label: 'Cards', icon: <LayoutGrid />, disabled: true },
      ]}
      initial="list"
      aria-label="View"
    />
  ),
  play: async ({ canvas, userEvent }) => {
    const cards = canvas.getByRole('radio', { name: 'Cards' })
    await expect(cards).toBeDisabled()
    // Arrows must skip it rather than parking selection on a dead segment.
    await userEvent.click(canvas.getByRole('radio', { name: 'List' }))
    await userEvent.keyboard('{ArrowRight}')
    await expect(canvas.getByRole('radio', { name: 'List' })).toHaveAttribute('aria-checked', 'true')
  },
}

export const LongLabelsDoNotWrap: Story = {
  render: () => (
    <div className="w-56">
      <Harness
        options={[
          { value: 'a', label: 'Everything, including archived' },
          { value: 'b', label: 'Only what needs attention' },
        ]}
        initial="a"
        aria-label="Scope"
      />
    </div>
  ),
}

/** Arrow keys move the selection and the focus with it; the group is a single
 *  Tab stop (roving tabindex), so Tab does not walk segment by segment. */
export const KeyboardNavigation: Story = {
  render: () => <Harness options={DENSITY} initial="compact" aria-label="Density" />,
  play: async ({ canvas, userEvent }) => {
    const [compact, cosy, roomy] = ['Compact', 'Cosy', 'Roomy'].map((n) =>
      canvas.getByRole('radio', { name: n }),
    )
    await expect(canvas.getByRole('radiogroup', { name: 'Density' })).toBeInTheDocument()

    await userEvent.tab()
    await expect(compact).toHaveFocus()

    await userEvent.keyboard('{ArrowRight}')
    await expect(cosy).toHaveFocus()
    await expect(cosy).toHaveAttribute('aria-checked', 'true')
    await expect(compact).toHaveAttribute('aria-checked', 'false')

    await userEvent.keyboard('{End}')
    await expect(roomy).toHaveAttribute('aria-checked', 'true')

    // Wraps rather than dead-ending.
    await userEvent.keyboard('{ArrowRight}')
    await expect(compact).toHaveAttribute('aria-checked', 'true')
  },
}

/** Selection is ink vs `fg-muted` — never opacity. A dimmed icon reads as
 *  disabled, which is the one state this control must never be confused with. */
export const SelectionIsInkNotOpacity: Story = {
  render: () => <Harness options={DENSITY} initial="compact" aria-label="Density" />,
  play: async ({ canvas }) => {
    const selected = canvas.getByRole('radio', { name: 'Compact' })
    const unselected = canvas.getByRole('radio', { name: 'Cosy' })

    for (const el of [selected, unselected]) {
      await expect(getComputedStyle(el).opacity).toBe('1')
    }
    await expect(getComputedStyle(selected).color).not.toBe(getComputedStyle(unselected).color)
  },
}

/** The one thing that makes a bordered segment inside a bordered track legible:
 *  the track has NO padding, so the two edges land on the same pixel instead of
 *  2px apart. At 2px the eye reads a doubled line rather than a raised surface. */
export const EdgesCoincideRatherThanDoubling: Story = {
  render: () => <Harness options={DENSITY} initial="compact" aria-label="Density" />,
  play: async ({ canvas }) => {
    const track = canvas.getByRole('radiogroup', { name: 'Density' })
    const first = canvas.getByRole('radio', { name: 'Compact' })
    const last = canvas.getByRole('radio', { name: 'Roomy' })

    for (const side of ['paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft'] as const) {
      await expect(parseFloat(getComputedStyle(track)[side])).toBe(0)
    }

    // The selected segment at the left end sits flush inside the track's own
    // border — one hairline between them, not two.
    const t = track.getBoundingClientRect()
    const f = first.getBoundingClientRect()
    const border = parseFloat(getComputedStyle(track).borderLeftWidth)
    await expect(Math.round(f.left - t.left)).toBe(Math.round(border))
    await expect(Math.round(f.top - t.top)).toBe(Math.round(border))

    // And the segments span the track edge to edge, leaving no track fill
    // showing past the last one.
    await expect(Math.round(t.right - last.getBoundingClientRect().right)).toBe(Math.round(border))
  },
}

/** A selected segment in the middle draws the divider with its own edge, so a
 *  separate rule can never end up doubled beside it. */
export const SelectedSegmentDrawsTheDivider: Story = {
  render: () => <Harness options={DENSITY} initial="cosy" aria-label="Density" />,
  play: async ({ canvas }) => {
    const middle = canvas.getByRole('radio', { name: 'Cosy' })
    const edge = canvas.getByRole('radio', { name: 'Compact' })
    const style = getComputedStyle(middle)

    await expect(parseFloat(style.borderLeftWidth)).toBe(1)
    await expect(parseFloat(style.borderRightWidth)).toBe(1)
    // An unselected segment carries none — the divider belongs to the selection.
    await expect(parseFloat(getComputedStyle(edge).borderLeftWidth)).toBe(0)
  },
}

/** §8 — the track takes its own height's radius, and the SELECTED segment
 *  repeats it on the outer corners it owns. */
export const RadiusTracksHeight: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Harness options={DENSITY} initial="compact" size="sm" aria-label="Small" />
      <Harness options={DENSITY} initial="compact" size="default" aria-label="Default" />
    </div>
  ),
  play: async ({ canvas }) => {
    const expected = { Small: [28, 6], Default: [32, 8] } as const
    for (const [name, [height, radius]] of Object.entries(expected)) {
      const track = canvas.getByRole('radiogroup', { name })
      const style = getComputedStyle(track)
      await expect(parseFloat(style.height)).toBe(height)
      await expect(parseFloat(style.borderRadius)).toBe(radius)
      await expect(style.overflow).toBe('hidden')

      // `compact` is selected and it is the FIRST segment, so it owns the two
      // left corners and neither right one. The track would clip them square
      // either way — the segment carries its own so that its shadow follows
      // the curve instead of cutting a corner across it.
      const segment = canvas.getAllByRole('radio').find((r) => track.contains(r))!
      const seg = getComputedStyle(segment)
      await expect(parseFloat(seg.borderTopLeftRadius)).toBe(radius)
      await expect(parseFloat(seg.borderBottomLeftRadius)).toBe(radius)
      await expect(parseFloat(seg.borderTopRightRadius)).toBe(0)
      await expect(parseFloat(seg.borderBottomRightRadius)).toBe(0)
    }
  },
}

/**
 * **The selected segment is raised** — the one piece of content in the product
 * allowed an elevation (§5). It is a key sitting proud of a keyboard, not a
 * dialog floating over a page, and the track clips the shadow on three sides so
 * what you see is a lift along the divider and nothing else.
 */
export const SelectedSegmentIsRaised: Story = {
  render: () => <Harness options={DENSITY} initial="cosy" aria-label="Density" />,
  play: async ({ canvas }) => {
    const selected = canvas.getByRole('radio', { name: 'Cosy' })
    const unselected = canvas.getByRole('radio', { name: 'Compact' })

    // The lift is `shadow-sm` and nothing heavier — a raised face, not a float.
    const shadow = getComputedStyle(selected).boxShadow
    await expect(shadow).not.toBe('none')
    await expect(getComputedStyle(unselected).boxShadow).toBe('none')

    // And the track clips it, so it can never spill onto the page.
    const track = canvas.getByRole('radiogroup', { name: 'Density' })
    await expect(getComputedStyle(track).overflow).toBe('hidden')
  },
}

/** The selected segment is the sheet's own white, and the track sits just
 *  behind it — a step, not a well. A darker track under a white segment would
 *  make the unselected side look switched off. */
export const SelectedSegmentComesForward: Story = {
  render: () => <Harness options={DENSITY} initial="compact" aria-label="Density" />,
  play: async ({ canvas }) => {
    const luminance = (rgb: string) => {
      const [r, g, b] = rgb.match(/\d+/g)!.map(Number)
      return 0.2126 * r + 0.7152 * g + 0.0722 * b
    }
    // Resolve --card through a real element so the value comes back as rgb()
    // in whichever theme is live, rather than as the raw hex in the token.
    const probe = document.createElement('div')
    probe.className = 'bg-card'
    document.body.appendChild(probe)
    const sheet = luminance(getComputedStyle(probe).backgroundColor)
    probe.remove()

    const track = luminance(
      getComputedStyle(canvas.getByRole('radiogroup', { name: 'Density' })).backgroundColor,
    )
    const thumb = luminance(
      getComputedStyle(canvas.getByRole('radio', { name: 'Compact' })).backgroundColor,
    )

    // The selected segment is nearer the sheet than the track is, in both
    // themes — the step must not invert when the palette does.
    await expect(Math.abs(thumb - sheet)).toBeLessThan(Math.abs(track - sheet))
  },
}
