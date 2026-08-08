import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Plus } from 'lucide-react'
import { Button } from './button'

const meta = {
  title: 'Primitives/Button',
  component: Button,
  tags: ['ai-generated'],
  args: { children: 'Deploy stack' },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Destructive: Story = { args: { variant: 'destructive', children: 'Delete cluster' } }
export const Outline: Story = { args: { variant: 'outline' } }
export const Secondary: Story = { args: { variant: 'secondary' } }
export const Ghost: Story = { args: { variant: 'ghost' } }
export const Link: Story = { args: { variant: 'link' } }
export const Inverse: Story = { args: { variant: 'inverse' } }
export const Mono: Story = { args: { variant: 'mono', children: 'Open console' } }
export const RailPrimary: Story = { args: { variant: 'railPrimary', size: 'rail', children: 'Save' } }
export const RailGhost: Story = { args: { variant: 'railGhost', size: 'rail', children: 'Cancel' } }
export const RailDanger: Story = { args: { variant: 'railDanger', size: 'rail', children: 'Discard' } }
export const Disabled: Story = {
  args: { disabled: true },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('button', { name: /deploy stack/i })).toBeDisabled()
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
  'mono',
  'railPrimary',
  'railGhost',
  'railDanger',
] as const

// Every exported variant, rest and disabled, in one view — the reference
// grid for eyeballing the pill/edge/press material across both themes.
export const VariantsGrid: Story = {
  render: () => (
    <div className="flex flex-col gap-4 bg-background p-6">
      <div className="flex flex-wrap items-center gap-3">
        {GRID_VARIANTS.map((variant) => (
          <Button key={variant} variant={variant} size={variant.startsWith('rail') ? 'rail' : 'default'}>
            {variant}
          </Button>
        ))}
      </div>
      <div className="flex flex-wrap items-center gap-3">
        {GRID_VARIANTS.map((variant) => (
          <Button
            key={variant}
            variant={variant}
            size={variant.startsWith('rail') ? 'rail' : 'default'}
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
        <Button key={variant} variant={variant} size={variant.startsWith('rail') ? 'rail' : 'default'}>
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
    await expect(style.outlineStyle).not.toBe('none')
    await expect(style.outlineWidth).toBe('2px')
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
