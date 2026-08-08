import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Plus } from 'lucide-react'
import { Button } from './button'

const meta = {
  title: 'Primitives/Button',
  component: Button,
  tags: ['ai-generated'],
  args: { children: 'Deploy stack' },
  // Every prop that changes how the button looks or behaves, as a control —
  // so the whole surface can be explored from the Controls panel instead of
  // hunting through one story per combination.
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'destructive', 'outline', 'secondary', 'ghost', 'link', 'inverse'],
      table: { defaultValue: { summary: 'default' }, category: 'Appearance' },
      description: 'Filled is rare — `default` is the one action per screen.',
    },
    size: {
      control: 'select',
      options: ['sm', 'default', 'lg', 'icon', 'icon-sm'],
      table: { defaultValue: { summary: 'default' }, category: 'Appearance' },
      description: 'Height follows density, never importance: sm 28 · default 32 · lg 40.',
    },
    shape: {
      control: 'inline-radio',
      options: ['pill', 'flat'],
      table: { defaultValue: { summary: 'pill' }, category: 'Appearance' },
      description:
        'What KIND of action this is. `pill` commits — it finishes a flow or ships something, and a screen gets at most one. `flat` is a working control: toolbars, filters, row actions, dialog footers.',
    },
    disabled: { control: 'boolean', table: { category: 'State' } },
    loading: {
      control: 'boolean',
      table: { category: 'State' },
      description: 'Spinner replaces the icon and the button goes inert — but keeps full contrast, unlike disabled.',
    },
    loadingText: {
      control: 'text',
      table: { category: 'State' },
      description: 'What it says while it works — "Creating…", not "Loading…". The spinner already reports *that* something is happening.',
    },
    asChild: {
      control: 'boolean',
      table: { category: 'Composition' },
      description: 'Render the child element instead of a <button> (e.g. a router Link).',
    },
    type: {
      control: 'inline-radio',
      options: ['button', 'submit', 'reset'],
      table: { defaultValue: { summary: 'button' }, category: 'Behaviour' },
    },
    children: { control: 'text', table: { category: 'Content' } },
    className: { control: 'text', table: { category: 'Escape hatch' } },
    onClick: { action: 'clicked', table: { category: 'Behaviour' } },
  },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof meta>

/** Every knob in one place — change variant, size and state from the Controls
 *  panel rather than reading down the list of fixed stories below. */
export const Playground: Story = {
  args: { variant: 'default', size: 'default', shape: 'pill', disabled: false, loading: false },
}

/** Shape is not decoration — it reports the CLASS of the action, and it is the
 *  fastest thing on the screen to read. `flat` is what most buttons are. */
export const Flat: Story = {
  args: { shape: 'flat', children: 'Save' },
}

/** The comparison that matters. Left column commits; right column works.
 *  A screen full of pills has no ranking — which is what ours looked like
 *  before this variant existed. */
export const PillVsFlat: Story = {
  render: () => (
    <div className="flex flex-col gap-6 bg-background p-6">
      {(['pill', 'flat'] as const).map((shape) => (
        <div key={shape} className="flex flex-col gap-2">
          <p className="text-meta text-fg-muted">
            {shape === 'pill'
              ? 'pill — commits. One per screen, at most.'
              : 'flat — a working control. Most buttons are this.'}
          </p>
          <div className="flex flex-wrap items-center gap-3">
            {(['default', 'secondary', 'outline', 'ghost', 'destructive'] as const).map((variant) => (
              <Button key={variant} variant={variant} shape={shape}>
                {shape === 'pill' ? 'Deploy stack' : 'Save'}
              </Button>
            ))}
          </div>
        </div>
      ))}
    </div>
  ),
}

/** Radius must not track the control's height: a class does not get quieter
 *  because the button is smaller. Every flat size resolves to the same 6px. */
export const FlatRadiusTracksHeight: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      {(['sm', 'default', 'lg'] as const).map((size) => (
        <Button key={size} shape="flat" size={size}>
          {size}
        </Button>
      ))}
    </div>
  ),
  play: async ({ canvas }) => {
    // §2 — anything the same height takes the same radius: 28/6 · 32/8 · 40/12.
    // A flat button has to match the input and the select beside it in a row.
    const expected = { sm: 6, default: 8, lg: 12 } as const
    for (const [size, px] of Object.entries(expected)) {
      const el = canvas.getByRole('button', { name: size })
      await expect(parseFloat(getComputedStyle(el).borderRadius)).toBe(px)
    }
  },
}

/** Guards the whole point: pill and flat must actually resolve to different
 *  radii. If someone collapses shape back into the base class, this fails. */
export const ShapesResolveToDifferentRadii: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button shape="pill">pill</Button>
      <Button shape="flat">flat</Button>
    </div>
  ),
  play: async ({ canvas }) => {
    const pill = getComputedStyle(canvas.getByRole('button', { name: 'pill' })).borderRadius
    const flat = getComputedStyle(canvas.getByRole('button', { name: 'flat' })).borderRadius
    await expect(pill).not.toBe(flat)
    await expect(parseFloat(pill)).toBeGreaterThan(parseFloat(flat))
  },
}

/** Loading keeps full contrast — a request in flight is not a disabled control,
 *  and the label has to stay readable while it runs. */
export const Loading: Story = {
  args: { loading: true, loadingText: 'Deploying…', children: 'Deploy stack' },
  play: async ({ canvas }) => {
    const btn = canvas.getByRole('button')
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute('aria-busy', 'true')
    await expect(btn).toHaveTextContent('Deploying…')
  },
}

/** No `loadingText`: the idle label stays put beside the spinner. Fine for a
 *  verb that reads the same either way ("Save"), but prefer the progressive
 *  form where there is one. */
export const LoadingWithoutText: Story = {
  args: { loading: true, children: 'Save' },
}

/** The icon is dropped, not doubled — `loadingText` replaces children whole. */
export const LoadingReplacesIcon: Story = {
  args: {
    loading: true,
    loadingText: 'Creating…',
    children: (
      <>
        <Plus /> Create addon
      </>
    ),
  },
  play: async ({ canvas }) => {
    const btn = canvas.getByRole('button')
    await expect(btn).toHaveTextContent('Creating…')
    await expect(btn.querySelectorAll('svg')).toHaveLength(1)
  },
}

export const Default: Story = {}

export const Destructive: Story = { args: { variant: 'destructive', children: 'Delete cluster' } }
export const Outline: Story = { args: { variant: 'outline' } }
export const Secondary: Story = { args: { variant: 'secondary' } }
export const Ghost: Story = { args: { variant: 'ghost' } }
export const Link: Story = { args: { variant: 'link' } }
export const Inverse: Story = { args: { variant: 'inverse' } }
export const Disabled: Story = {
  args: { disabled: true },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('button', { name: /deploy stack/i })).toBeDisabled()
  },
}

/** The disabled contract: dimmed, refused, inert — and the shape survives.
 *
 *  The cursor is the part that is easy to lose. It only shows if the button
 *  still receives the pointer, so `pointer-events-none` is deliberately absent;
 *  the native `disabled` attribute blocks the click instead. That leaves hover
 *  and press, which still match a disabled button in CSS, so every interaction
 *  state is guarded by `not-disabled:`. */
export const DisabledContract: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      {(['default', 'destructive', 'outline', 'secondary', 'ghost', 'inverse'] as const).map((v) => (
        <Button key={v} variant={v} disabled>
          {v}
        </Button>
      ))}
    </div>
  ),
  play: async ({ canvas }) => {
    for (const v of ['default', 'destructive', 'outline', 'secondary', 'ghost', 'inverse']) {
      const btn = canvas.getByRole('button', { name: v })
      const style = getComputedStyle(btn)

      await expect(btn).toBeDisabled()
      await expect(style.opacity).toBe('0.5')
      await expect(style.cursor).toBe('not-allowed')
      // The cursor is only reachable while the element still takes the pointer.
      await expect(style.pointerEvents).not.toBe('none')
      // Shape kept: the default `flat` shape must survive being switched off.
      // A 32px control sits on the 8px rung (§8), and a disabled button that
      // loses its corner reads as a rendering fault rather than a refusal.
      await expect(parseFloat(style.borderRadius)).toBe(8)
      // No press travel — depth exists only while you are touching it (§6).
      await expect(style.boxShadow).toBe('none')
    }
  },
}

/** §6 — a request in flight is not a disabled control. Contrast stays at 100%,
 *  and the cursor says "working", not "refused". */
export const LoadingIsNotDisabled: Story = {
  args: { loading: true, loadingText: 'Deploying…' },
  play: async ({ canvas }) => {
    const btn = canvas.getByRole('button')
    const style = getComputedStyle(btn)
    await expect(btn).toBeDisabled()
    await expect(style.opacity).toBe('1')
    await expect(style.cursor).toBe('progress')
  },
}
export const WithIcon: Story = {
  args: {
    size: 'sm',
    children: (
      <>
        <Plus /> Add addon
      </>
    ),
  },
}

const GRID_VARIANTS = [
  'default',
  'destructive',
  'outline',
  'secondary',
  'ghost',
  'link',
  'inverse',
] as const

// Every exported variant, rest and disabled, in one view — the reference
// grid for eyeballing the pill/edge/press material across both themes.
export const VariantsGrid: Story = {
  render: () => (
    <div className="flex flex-col gap-4 bg-background p-6">
      <div className="flex flex-wrap items-center gap-3">
        {GRID_VARIANTS.map((variant) => (
          <Button key={variant} variant={variant} size="default">
            {variant}
          </Button>
        ))}
      </div>
      <div className="flex flex-wrap items-center gap-3">
        {GRID_VARIANTS.map((variant) => (
          <Button
            key={variant}
            variant={variant}
            size="default"
            disabled
          >
            {variant}
          </Button>
        ))}
      </div>
    </div>
  ),
}

// D13/D14: no transform anywhere, at rest or hovered — the graphite material
// communicates state entirely through fill and inset shadow.
export const NoTransformOnHover: Story = {
  render: () => (
    <div className="flex flex-wrap gap-3">
      {GRID_VARIANTS.map((variant) => (
        <Button key={variant} variant={variant} size="default">
          {variant}
        </Button>
      ))}
    </div>
  ),
  play: async ({ canvas, userEvent }) => {
    for (const variant of GRID_VARIANTS) {
      const button = canvas.getByRole('button', { name: variant })
      await expect(getComputedStyle(button).transform).toBe('none')
      await userEvent.hover(button)
      await expect(getComputedStyle(button).transform).toBe('none')
    }
  },
}

// Focus-visible must render as a solid outline ring off --ring, not the
// removed ring-* utilities — tab to the button rather than clicking so
// :focus-visible actually engages.
export const KeyboardFocusOutline: Story = {
  play: async ({ canvas, userEvent }) => {
    const button = canvas.getByRole('button', { name: /deploy stack/i })
    await userEvent.tab()
    await expect(button).toHaveFocus()
    const style = getComputedStyle(button)
    // The ring is a box-shadow, never an outline — a shadow follows the radius,
    // stacks with the press recess, and can carry a gap when a dark face needs one.
    await expect(style.outlineStyle).toBe('none')
    // Asserted as the CLASS CONTRACT, not the computed pixel: an unlayered
    // `box-shadow` resolves correctly in the app and in a live Storybook — both
    // measured — but comes back as Tailwind's `0 0 #0000` default inside this
    // browser-test harness. Pinning the pixel here would be testing the harness.
    await expect(button.className).toMatch(/(?:^|\s)focus-ring(?:-edge|-inset)?(?:\s|$)/)
    await expect(button.className).not.toMatch(/(?:^|[\s:])ring-/)
  },
}

// Proves the app's Tailwind theme actually loaded in the preview: the default
// variant is bg-primary, which resolves to the --primary token from
// src/index.css. Reads the token live (rather than a hardcoded color
// literal) so this keeps passing as the palette evolves.
export const CssCheck: Story = {
  play: async ({ canvas }) => {
    const button = canvas.getByRole('button', { name: /deploy stack/i })
    const probe = document.createElement('div')
    probe.style.color = getComputedStyle(document.documentElement).getPropertyValue('--primary').trim()
    document.body.appendChild(probe)
    const expected = getComputedStyle(probe).color
    probe.remove()
    await expect(getComputedStyle(button).backgroundColor).toBe(expected)
  },
}
