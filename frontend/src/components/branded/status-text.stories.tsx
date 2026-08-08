import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { StatusText } from './status-text'

/** The closed rollup vocabulary, mirrored from `StackRollupState`. */
const ROLLUP_STATES = [
  'Healthy',
  'Deploying',
  'Degraded',
  'Unavailable',
  'Failed',
  'NotDeployed',
  'Deleting',
] as const

const meta = {
  title: 'Branded/StatusText',
  component: StatusText,
  args: { domain: 'stack', state: 'Ready' },
} satisfies Meta<typeof StatusText>

export default meta
type Story = StoryObj<typeof meta>

export const Ready: Story = {}
export const Pending: Story = { args: { state: 'Progressing' } }
export const Failed: Story = { args: { state: 'Failed' } }
export const Degraded: Story = { args: { state: 'Degraded' } }

/** An unrecognised word is shown, not swallowed. It must never come back green:
 *  a silent "ready" for a state nobody mapped is the failure mode that hides
 *  a broken deploy. */
export const UnknownWord: Story = { args: { state: 'Quiescing' } }

/** Missing state is neutral, and says so rather than rendering an empty cell. */
export const NoState: Story = { args: { state: null } }

export const EveryStackState: Story = {
  render: () => (
    <div className="flex flex-col gap-1">
      {['Ready', 'Pending', 'Progressing', 'Deleting', 'Failed', 'Degraded', 'Error'].map((s) => (
        <StatusText key={s} domain="stack" state={s} />
      ))}
    </div>
  ),
}

/** Machine casing is made readable without losing which word it was —
 *  `Degraded` and `Failed` are both "error", and which one you are looking at
 *  changes what you do next, so the bucket label is not what renders. */
export const MachineCasingIsHumanised: Story = {
  render: () => (
    <div className="flex flex-col gap-1">
      <StatusText domain="release" state="InProgress" />
      <StatusText domain="generic" state="image_pull_failed" />
      <StatusText domain="generic" state="crashloopbackoff" />
    </div>
  ),
  play: async ({ canvas }) => {
    await expect(canvas.getByText('In progress')).toBeInTheDocument()
    await expect(canvas.getByText('Image pull failed')).toBeInTheDocument()
  },
}

/**
 * THE CONTRACT. The defect this component exists to remove is a red row that
 * reads "Healthy" — which was reachable because `variant` and the word were two
 * independent props.
 *
 * This asserts the contradiction is *unbuildable*, not merely discouraged:
 * there is no prop that carries a colour, no prop that carries a word, and no
 * `className` to smuggle one in through. The colour is a pure function of the
 * state, so a caller cannot desynchronise them without editing this file.
 */
export const ContradictionIsUnbuildable: Story = {
  render: () => {
    // The escape hatches a caller would reach for, forced past the type system.
    // TypeScript rejects all three; this proves the runtime rejects them too,
    // so the guarantee does not rest on nobody writing `as any`.
    const Forced = StatusText as unknown as (p: Record<string, unknown>) => React.ReactElement
    return (
      <div className="flex flex-col gap-1">
        <StatusText domain="stack" state="Failed" />
        <StatusText domain="stack" state="Ready" />
        <Forced domain="stack" state="Failed" className="text-success" />
        <Forced domain="stack" state="Failed" style={{ color: 'green' }} />
        <Forced domain="stack" state="Failed" children="Healthy" />
      </div>
    )
  },
  play: async ({ canvas }) => {
    const ready = canvas.getByText('Ready')
    const failures = canvas.getAllByText('Failed')

    // 1. The word and the colour come from the same input, so they agree.
    await expect(failures[0]).toHaveAttribute('data-status-variant', 'error')
    await expect(ready).toHaveAttribute('data-status-variant', 'ready')
    await expect(getComputedStyle(failures[0]).color).not.toBe(getComputedStyle(ready).color)

    // 2. The word "Healthy" was passed as children and does not exist on screen.
    //    A red row cannot be made to read Healthy.
    await expect(canvas.queryByText('Healthy')).toBeNull()

    // 3. Every forced variant renders identically to the honest one — the
    //    className, the inline style and the children all fell on the floor.
    await expect(failures).toHaveLength(4)
    const honest = getComputedStyle(failures[0]).color
    for (const el of failures) {
      await expect(getComputedStyle(el).color).toBe(honest)
      await expect(el.getAttribute('style')).toBeNull()
      await expect(el.textContent).toBe('Failed')
    }

    // 4. That colour is the token for the derived variant — not a literal
    //    someone chose. Probed live so it keeps holding as the palette moves.
    const probe = document.createElement('span')
    probe.className = 'text-danger'
    document.body.appendChild(probe)
    const danger = getComputedStyle(probe).color
    probe.remove()
    await expect(honest).toBe(danger)
  },
}

/** §7 — status is said once. No dot: a coloured dot plus the word in a column
 *  is saying it twice, and lists are where that slips. */
export const SaysItOnce: Story = {
  args: { state: 'Failed' },
  play: async ({ canvas }) => {
    const el = canvas.getByText('Failed')
    await expect(el.children).toHaveLength(0)
    await expect(el.textContent).toBe('Failed')
    // No chip either — no border, no fill. The colour is the whole signal.
    const style = getComputedStyle(el)
    await expect(parseFloat(style.borderTopWidth)).toBe(0)
    await expect(style.backgroundColor).toBe('rgba(0, 0, 0, 0)')
  },
}

export const LongUnknownStateDoesNotWrap: Story = {
  render: () => (
    <div className="w-24 border border-border p-2">
      <StatusText domain="generic" state="create_container_error" />
    </div>
  ),
}

/**
 * **One glyph per state, not per family.** The set used to be three icons for
 * three families, so `Degraded`, `Unavailable` and `Failed` all drew the same
 * triangle — a mark that could not make the distinction the word already made.
 * Both list views had icons switched off for exactly that reason.
 *
 * This is the story that holds the set honest: every rollup state, side by
 * side, and the assertion below is that no two of them share a glyph.
 */
export const EveryRollupStateHasItsOwnGlyph: Story = {
  render: () => (
    <div className="flex flex-col gap-2">
      {ROLLUP_STATES.map((s) => (
        <StatusText key={s} domain="stack_rollup" state={s} icon />
      ))}
    </div>
  ),
  play: async ({ canvas }) => {
    const shapes = ROLLUP_STATES.map((s) => {
      const word = s.replace(/([a-z])([A-Z])/g, '$1 $2')
      const label = word.charAt(0).toUpperCase() + word.slice(1).toLowerCase()
      const svg = canvas.getByText(label).querySelector('svg')
      // The rendered path data IS the glyph identity — comparing icon component
      // names would only re-assert the map, not what the user sees.
      return [...(svg?.querySelectorAll('path, circle, line, rect, polyline') ?? [])]
        .map((n) => n.outerHTML)
        .join('')
    })
    await expect(shapes.every((s) => s.length > 0)).toBe(true)
    await expect(new Set(shapes).size).toBe(ROLLUP_STATES.length)
  },
}

/** The in-flight state is the only one that moves, and it is `motion-safe:` so
 *  reduced-motion still gets the mark — just still. */
export const DeployingSpins: Story = {
  args: { domain: 'stack_rollup', state: 'Deploying', icon: true },
  play: async ({ canvas }) => {
    const svg = canvas.getByText('Deploying').querySelector('svg')
    await expect(svg?.getAttribute('class')).toContain('motion-safe:animate-spin')
  },
}

/** Icons are opt-in. Without the flag the word stands alone, which is what
 *  every other status surface in the product still does. */
export const IconIsOptIn: Story = {
  args: { domain: 'stack_rollup', state: 'Healthy' },
  play: async ({ canvas }) => {
    await expect(canvas.getByText('Healthy').querySelector('svg')).toBeNull()
  },
}
