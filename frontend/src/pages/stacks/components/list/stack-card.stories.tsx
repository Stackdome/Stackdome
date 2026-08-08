import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { makeStack } from '../../../../../.storybook/fixtures'
import { ReleaseState } from '@/pages/stacks/components/editor/tabs/deployments/release-states'
import type { Stack } from '@/api/stack-types'
import { DeployStackCard, StackCardSkeleton } from './stack-card'

const meta = {
  title: 'Features/StackCard',
  component: DeployStackCard,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      // The real three-up column at 1440: 1186 sheet − 24 padding − 24 gutter,
      // over three. The chip row is budgeted against exactly this width.
      <div className="w-[379px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof DeployStackCard>

export default meta
type Story = StoryObj<typeof meta>

const released = {
  latest_release: { id: 'r1', state: ReleaseState.Released },
  converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
} as Partial<Stack>

// §7's card material: white fill, 1px hairline, `rounded-lg`, no shadow. The
// whole card navigates, and hover is a wash plus the stronger hairline — no
// lift, no scale, and never a brand-coloured ring.
export const Running: Story = {
  args: { stack: makeStack(released), onDelete: fn() },
  play: async ({ canvas, canvasElement }) => {
    const card = canvas.getByRole('link')
    const style = getComputedStyle(card)

    await expect(style.boxShadow).toBe('none')
    await expect(parseFloat(style.borderTopWidth)).toBe(1)
    await expect(parseFloat(style.borderRadius)).toBe(12)

    // White, same as the sheet — grey never means "a card" (§1).
    const probe = document.createElement('div')
    probe.className = 'bg-card'
    document.body.appendChild(probe)
    await expect(style.backgroundColor).toBe(getComputedStyle(probe).backgroundColor)
    probe.remove()

    await expect(card.className).toContain('hover:border-border-strong')
    await expect(card.className).toContain('focus-ring')
    // No Tailwind `ring-*` utility — the focus ring is a box-shadow. Anchored so
    // it cannot trip on the `--focus-ring-edge` variable in the shadow above.
    await expect(card.className).not.toMatch(/(?:^|[\s:])ring-/)
    await expect(canvasElement.querySelector('svg.text-brand')).toBeNull()
  },
}

/**
 * **Every card is 162px, whatever it has to say.**
 *
 * The content needs 144. The remaining 18 sits between the head group and the
 * footer, so the fullest card breathes rather than the emptiest one sitting
 * short — otherwise one broken stack shoves the whole grid around on every poll.
 *
 * The gaps, not just the boxes: 2px name→provenance, 8px reason→chips, and the
 * footer on the card's own floor with no rule above it.
 */
export const FixedHeight: Story = {
  args: { stack: makeStack(released), onDelete: fn() },
  play: async ({ canvas, canvasElement }) => {
    const card = canvas.getByRole('link')
    const box = card.getBoundingClientRect()
    await expect(box.height).toBe(162)

    const at = (slot: string) =>
      canvasElement.querySelector(`[data-slot='${slot}']`)!.getBoundingClientRect()
    const name = canvas.getByText('orders-api').getBoundingClientRect()
    const prov = at('provenance')
    const chips = at('components')
    const footer = at('card-footer')

    await expect(Math.round(prov.top - name.bottom)).toBe(2)
    await expect(Math.round(chips.top - prov.bottom)).toBe(8)
    // The footer sits on the card's own floor: 16px padding plus the 1px
    // hairline, which `getBoundingClientRect` counts.
    await expect(Math.round(box.bottom - footer.bottom)).toBe(17)
    await expect(Math.round(footer.height)).toBe(16)
    // Healthy carries no reason line, so it hands that row AND the board's
    // slack to the gap above the footer: 24 + 16 reason + 2 = 42.
    await expect(Math.round(footer.top - chips.bottom)).toBe(42)
  },
}

/**
 * The **fullest** state — a reason line on top of everything else — is still
 * 162, and still has air above its footer. That is the whole test of the
 * height: if the tallest card were the one that fits exactly, every other
 * state would be sitting short of a floor it can never reach.
 */
export const FullestStateStillFits: Story = {
  args: {
    stack: makeStack({
      name: 'checkout-api',
      latest_release: { id: 'r1', state: ReleaseState.Released },
      converged_release: {
        id: 'r1',
        state: ReleaseState.Released,
        health: 'degraded',
        message: 'session · 1 of 3 replicas available',
      },
    } as Partial<Stack>),
    onDelete: fn(),
  },
  play: async ({ canvas, canvasElement }) => {
    await expect(canvas.getByRole('link').getBoundingClientRect().height).toBe(162)
    const at = (slot: string) =>
      canvasElement.querySelector(`[data-slot='${slot}']`)!.getBoundingClientRect()
    // 160 inside the hairline − 32 padding = 128 for the content. The fullest
    // head is 88 and the footer 16, so 24 of air is left over — the tallest
    // state still breathes rather than fitting exactly.
    await expect(Math.round(at('card-footer').top - at('components').bottom)).toBe(24)
    await expect(canvas.getByText('session · 1 of 3 replicas available')).toBeVisible()
  },
}

/**
 * **No rule above the footer.** A hairline across the middle of a 162px box
 * reads as two stacked cards; the 18px of air and the card's own edge already
 * say where the footer starts.
 */
export const FooterHasNoRule: Story = {
  args: { stack: makeStack(released), onDelete: fn() },
  play: async ({ canvas, canvasElement }) => {
    const footer = canvas.getByText(/^Last change /).parentElement!
    await expect(getComputedStyle(footer).borderTopWidth).toBe('0px')
    await expect(canvasElement.querySelector('.border-t')).toBeNull()
  },
}

/**
 * The name is **SemiBold 16/24** and the status word **Medium 12/16**, and they
 * sit on **one baseline** — not centred against each other, which reads as two
 * boxes of different heights rather than one written line.
 */
export const NameAndStatusShareABaseline: Story = {
  args: { stack: makeStack(released) },
  play: async ({ canvas }) => {
    const name = canvas.getByText('orders-api')
    const word = canvas.getByText('Healthy')
    const nameStyle = getComputedStyle(name)

    await expect(nameStyle.fontSize).toBe('16px')
    await expect(nameStyle.fontWeight).toBe('600')
    await expect(getComputedStyle(word).fontSize).toBe('12px')

    // The mechanic, asserted where it lives — geometry alone cannot prove this
    // one: centring a 16px line in a 24px row lands its bottom 4px short of
    // the big line's, and sharing a baseline lands it 3px short. A single
    // pixel apart is not a test.
    await expect(getComputedStyle(name.parentElement!).alignItems).toBe('baseline')
    await expect(Math.round(name.getBoundingClientRect().bottom - word.getBoundingClientRect().bottom)).toBe(3)
  },
}

/**
 * Status is a **glyph and a word**, in one hue.
 *
 * The glyph was off here while the set was per-FAMILY: one triangle stood for
 * `Degraded`, `Unavailable` and `Failed`, so it added a symbol without adding a
 * fact. The set is per-STATE now, so the mark carries its own information.
 *
 * §7 still binds — what "says it once" forbids is a second *colour channel*.
 * No dot, no rail, no fill, no border: the glyph and the word share one tone
 * and there is nothing else reporting the same thing.
 */
export const StatusIsAGlyphAndAWord: Story = {
  args: { stack: makeStack(released) },
  play: async ({ canvas }) => {
    const word = canvas.getByText('Healthy')
    const svg = word.querySelector('svg')
    await expect(svg).not.toBeNull()
    // One colour channel: the glyph inherits the word's tone rather than
    // carrying a second one of its own.
    await expect(getComputedStyle(svg!).color).toBe(getComputedStyle(word).color)
    const style = getComputedStyle(word)
    await expect(parseFloat(style.borderTopWidth)).toBe(0)
    await expect(style.backgroundColor).toBe('rgba(0, 0, 0, 0)')
  },
}

// In flight reads amber, never brand orange (§4 — hue reports state). The
// status rail is gone: the word carries the colour, and it carries it once.
export const Deploying: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r2', state: ReleaseState.InProgress },
      converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
    } as Partial<Stack>),
  },
  play: async ({ canvas, canvasElement }) => {
    const word = canvas.getByText('Deploying')
    await expect(word).toHaveAttribute('data-status-variant', 'pending')

    const probe = document.createElement('span')
    probe.className = 'text-warn'
    document.body.appendChild(probe)
    await expect(getComputedStyle(word).color).toBe(getComputedStyle(probe).color)
    probe.remove()

    // Said once: no rail, no dot, no brand fill anywhere on the card.
    await expect(canvasElement.querySelector('[data-rail]')).toBeNull()
    await expect(canvasElement.querySelector('.bg-brand')).toBeNull()
  },
}

export const Degraded: Story = {
  args: {
    stack: makeStack({
      name: 'auth-gateway',
      latest_release: { id: 'r1', state: ReleaseState.Released },
      converged_release: {
        id: 'r1',
        state: ReleaseState.Released,
        health: 'degraded',
        message: 'session · 1 of 3 replicas available',
      },
    } as Partial<Stack>),
    onDelete: fn(),
  },
}

export const Unavailable: Story = {
  args: {
    stack: makeStack({
      name: 'checkout-api',
      latest_release: { id: 'r1', state: ReleaseState.Released },
      converged_release: {
        id: 'r1',
        state: ReleaseState.Released,
        health: 'unavailable',
        message: 'no replicas available',
      },
    } as Partial<Stack>),
    onDelete: fn(),
  },
}

export const Failed: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r1', state: ReleaseState.Failed },
    } as Partial<Stack>),
  },
}

// The latest push failed while a previous release is still live. These are two
// different facts and the card says both: the stack really is serving, so the
// status word stays "Healthy" — overwriting it with "Failed" would be a lie —
// and the failed push gets its own line rather than a bare icon. §4: if a
// colour doesn't report something it's a bug, and an unlabelled triangle makes
// the reader guess what it reports.
export const DeployFailedWhileLive: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r2', state: ReleaseState.Failed },
      converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
    } as Partial<Stack>),
  },
  play: async ({ canvas }) => {
    await expect(canvas.getByText('Healthy')).toHaveAttribute('data-status-variant', 'ready')
    const line = canvas.getByText('Last deploy failed')
    await expect(line).toBeVisible()

    // **The reason keeps the danger hue only when it CONTRADICTS the word above
    // it.** This is that case: green "Healthy" and a failed push are two true
    // facts, and the red is the only thing telling you the second one exists.
    const probe = document.createElement('span')
    probe.className = 'text-danger'
    document.body.appendChild(probe)
    await expect(getComputedStyle(line).color).toBe(getComputedStyle(probe).color)
    probe.remove()
  },
}

/**
 * The mirror of the case above: a **Degraded** card's own message is already
 * accounted for by the word beside it, so the line stays `fg-2` and quiet.
 * Colouring it red would spend the hue on a fact the status already reports —
 * and leave nothing louder for the contradiction that matters.
 */
export const ReasonIsQuietWhenTheStatusAlreadySaysIt: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r1', state: ReleaseState.Released },
      converged_release: {
        id: 'r1',
        state: ReleaseState.Released,
        health: 'degraded',
        message: '1 of 3 replicas available',
      },
    } as Partial<Stack>),
  },
  play: async ({ canvas }) => {
    const line = canvas.getByText('1 of 3 replicas available')
    const probe = document.createElement('span')
    probe.className = 'text-fg-2'
    document.body.appendChild(probe)
    await expect(getComputedStyle(line).color).toBe(getComputedStyle(probe).color)
    probe.remove()
  },
}

/**
 * **Two typefaces on one line (§6).** The project is a word and the ref is a
 * machine value you might paste into a terminal, so only the ref is mono.
 * Setting the whole line in mono — which is what the code did — made the
 * project name look like something you could type at a shell.
 */
export const ProvenanceMixesGeistAndMono: Story = {
  args: {
    stack: makeStack({
      ...released,
      spec: {
        stack_resources: [
          {
            name: 'web',
            workload_type: 'Service',
            source: {
              git: {
                repo_url: 'https://github.com/acme/orders',
                branch: 'main',
                commit: 'a1b2c3d4e5f6',
                dockerfile_path: 'Dockerfile',
                build_context: '.',
              },
            },
          },
        ],
        volumes: [],
      },
    } as Partial<Stack>),
    projectName: 'acme',
  },
  play: async ({ canvasElement }) => {
    const line = canvasElement.querySelector("[data-slot='provenance']")!
    await expect(line.textContent).toBe('acme · main@a1b2c3d')

    const mono = line.querySelector('span')!
    await expect(mono.textContent).toBe('main@a1b2c3d')
    // The ref is mono; the line it sits on is not.
    await expect(getComputedStyle(mono).fontFamily).toMatch(/mono/i)
    await expect(getComputedStyle(line).fontFamily).not.toMatch(/mono/i)
  },
}

export const NotDeployed: Story = {
  args: { stack: makeStack() },
}

export const Deleting: Story = {
  args: { stack: makeStack({ ...released, lifecycle: 'deleting' } as Partial<Stack>), onDelete: fn() },
  play: async ({ canvas }) => {
    // A stack on its way out is in flight, not in trouble: no reason line, and
    // the trash is inert rather than absent so the card keeps its shape.
    await expect(canvas.getByText('Deleting')).toBeVisible()
    await expect(canvas.getByLabelText(/^Delete /)).toBeDisabled()
  },
}

/**
 * A name longer than the card **truncates**; it never wraps to a second line
 * and never pushes the status word off the right edge. The card's height is
 * fixed, so a wrapped name would eat the provenance line.
 */
export const LongName: Story = {
  args: {
    stack: makeStack({
      ...released,
      name: 'extremely-long-service-name-that-truncates-in-the-card-header',
    } as Partial<Stack>),
  },
  play: async ({ canvas }) => {
    const card = canvas.getByRole('link')
    const name = canvas.getByText('extremely-long-service-name-that-truncates-in-the-card-header')
    const style = getComputedStyle(name)

    await expect(style.textOverflow).toBe('ellipsis')
    await expect(style.whiteSpace).toBe('nowrap')
    await expect(name.getBoundingClientRect().height).toBe(24)
    await expect(card.getBoundingClientRect().height).toBe(162)

    // The status word survives, on the card and to the right of the name.
    const word = canvas.getByText('Healthy').getBoundingClientRect()
    await expect(word.right).toBeLessThanOrEqual(card.getBoundingClientRect().right)
    await expect(word.left).toBeGreaterThan(name.getBoundingClientRect().left)
  },
}

export const GitSource: Story = {
  args: {
    stack: makeStack({
      ...released,
      spec: {
        stack_resources: [
          {
            name: 'web',
            workload_type: 'Service',
            source: {
              git: {
                repo_url: 'https://github.com/acme/orders',
                branch: 'main',
                dockerfile_path: 'Dockerfile',
                build_context: '.',
              },
            },
          },
        ],
        volumes: [],
      },
    } as Partial<Stack>),
  },
}

// No onDelete wired → no destructive control at all (viewer without write
// access). Absent, not disabled: there is nothing here for them to be denied.
export const ReadOnly: Story = {
  args: { stack: makeStack(released) },
  play: async ({ canvas }) => {
    await expect(canvas.queryByLabelText(/^Delete /)).toBeNull()
  },
}

/**
 * **Hover is a wash and a line, and it appears at rest as neither.**
 *
 * `:hover` cannot be driven from a play function, so this asserts the *rules*
 * the card carries and the at-rest state they resolve from: the trash is
 * invisible until the pointer arrives, and the card never grows a shadow,
 * a lift or a scale to announce itself. §11 — content is flat.
 */
export const HoverIsAWashNotALift: Story = {
  args: { stack: makeStack(released), onDelete: fn() },
  play: async ({ canvas }) => {
    const card = canvas.getByRole('link')
    await expect(card.className).toContain('hover:bg-[var(--wash-hover)]')
    await expect(card.className).toContain('hover:border-border-strong')
    // Nothing that moves the card.
    await expect(card.className).not.toContain('hover:shadow')
    await expect(card.className).not.toContain('hover:scale')
    await expect(card.className).not.toContain('-translate-y')
    // At rest the trash is not on screen; it still holds its tab stop.
    const trash = canvas.getByLabelText('Delete orders-api')
    await expect(parseFloat(getComputedStyle(trash).opacity)).toBe(0)
    await expect(trash.className).toContain('group-hover/card:opacity-100')
  },
}

/**
 * The card carries **no chart**, and `deploy_history` being present must not
 * conjure one.
 *
 * A fortnight of deploy volume was tried here and cut: fourteen bars with no
 * axis, no baseline and no unit cannot distinguish eleven deploys from four, so
 * it reported "there has been activity" and nothing anyone acts on. It drew the
 * *shape* of substance without supplying any.
 */
export const HistoryDrawsNoChart: Story = {
  args: {
    stack: makeStack({
      ...released,
      deploy_history: [3, 1, 0, 4, 2, 0, 0, 6, 1, 2, 0, 3, 5, 1],
    } as Partial<Stack>),
    onDelete: fn(),
  },
  play: async ({ canvas, canvasElement }) => {
    await expect(canvas.queryByText(/deploys/i)).toBeNull()
    await expect(canvasElement.querySelector("[data-slot='deploy-sparkline']")).toBeNull()
  },
}

/**
 * What the card carries instead: every component **by name**, and nothing else.
 *
 * `4 svc · 2 vol` is the stack's shape written as a count — it cannot be
 * recognised at a glance and there is nothing to do with it. A new joiner
 * reading `web · worker · uploads` learns what the thing IS, which is the whole
 * reason the cards view exists (§11). The **"N components" label above the
 * chips is gone too**, for the same reason: the chips already say how many.
 */
export const ComponentsByName: Story = {
  args: {
    stack: makeStack({
      ...released,
      spec: {
        stack_resources: [{ name: 'web' }, { name: 'worker' }],
        volumes: [{ name: 'uploads' }],
      },
    } as Partial<Stack>),
  },
  play: async ({ canvas, canvasElement }) => {
    for (const name of ['web', 'worker', 'uploads']) {
      await expect(canvas.getByText(name)).toBeInTheDocument()
    }
    await expect(canvas.queryByText(/\d+ components?/)).toBeNull()

    // §8's chip rung, measured: 20 high, 6 radius, 7px of side padding, 5 apart.
    const row = canvasElement.querySelector("[data-slot='components']")!
    const chips = [...row.children] as HTMLElement[]
    for (const chip of chips) {
      const s = getComputedStyle(chip)
      await expect(chip.getBoundingClientRect().height).toBe(20)
      await expect(parseFloat(s.borderRadius)).toBe(6)
      await expect(parseFloat(s.paddingLeft)).toBe(7)
      await expect(parseFloat(s.paddingRight)).toBe(7)
      await expect(parseFloat(s.borderTopWidth)).toBe(1)
    }
    const [a, b] = chips.map((c) => c.getBoundingClientRect())
    await expect(Math.round(b.left - a.right)).toBe(5)

    // The glyph stays: it reports the KIND, which is the one job §7 lets an
    // icon do. Without it `orders-db` and `uploads` render identically when one
    // is a database and one is a volume.
    for (const chip of chips) {
      await expect(chip.querySelector('svg')).not.toBeNull()
    }
  },
}

/**
 * More components than the row can hold: the overflow is **counted**, never
 * wrapped. The card's height is fixed, so a second chip row would either clip
 * silently or push the footer off the bottom.
 */
export const ChipsOverflowToACount: Story = {
  args: {
    stack: makeStack({
      ...released,
      spec: {
        stack_resources: [
          { name: 'web' },
          { name: 'worker' },
          { name: 'scheduler' },
          { name: 'orders-db' },
          { name: 'search-index' },
        ],
        volumes: [{ name: 'uploads' }],
      },
    } as Partial<Stack>),
  },
  play: async ({ canvas, canvasElement }) => {
    const more = canvas.getByText(/^\+\d+$/)
    await expect(more).toBeVisible()

    // A count is a word, not a machine value (§6) — so `+N` is Geist where the
    // component names beside it are mono.
    await expect(getComputedStyle(more).fontFamily).not.toMatch(/mono/i)
    await expect(getComputedStyle(canvas.getByText('web')).fontFamily).toMatch(/mono/i)

    // One row, and it stays one row: the card's height is fixed, so a second
    // line would push the footer off the bottom rather than wrap.
    const row = canvasElement.querySelector("[data-slot='components']")!
    await expect(row.getBoundingClientRect().height).toBe(20)
    await expect(canvas.getByRole('link').getBoundingClientRect().height).toBe(162)
  },
}

/**
 * The reason line gets **one line here, and it truncates.**
 *
 * That is the cost of the fixed height, and it is deliberate: the card's
 * advantage over the row is the component list, not the reason. The row's
 * status column is 504px wide where this one is 344 — so the *row* is where a
 * failure finishes its sentence.
 */
export const ReasonTruncatesToOneLine: Story = {
  args: {
    stack: makeStack({
      latest_release: {
        id: 'r9',
        state: ReleaseState.Failed,
        message: 'web · image pull failed — ghcr.io/acme/docs:e91a02 not found',
        completed_at: '2026-08-05T11:40:00Z',
      },
    } as Partial<Stack>),
  },
  play: async ({ canvas }) => {
    const line = canvas.getByText('web · image pull failed — ghcr.io/acme/docs:e91a02 not found')
    const style = getComputedStyle(line)
    await expect(style.textOverflow).toBe('ellipsis')
    await expect(style.whiteSpace).toBe('nowrap')
    await expect(line.getBoundingClientRect().height).toBe(16)
  },
}

/**
 * **The loading state is on the real card's rhythm**, block for block: 162
 * tall, the name / provenance / chips stacked in the head, and the footer on
 * the floor. A skeleton at a different height moves every card on the page the
 * moment the data lands, which is the one thing a loading state exists to
 * prevent. No shimmer (§14).
 */
export const Skeleton: StoryObj = {
  render: () => <StackCardSkeleton />,
  decorators: [
    (Story) => (
      <div className="w-[1186px]">
        <Story />
      </div>
    ),
  ],
  play: async ({ canvasElement }) => {
    const cards = [...canvasElement.querySelectorAll('[aria-hidden] > div')] as HTMLElement[]
    await expect(cards.length).toBe(6)
    for (const card of cards) {
      await expect(card.getBoundingClientRect().height).toBe(162)
    }
    // Same three rows as the real head, at the real gaps.
    const [name, prov, chips] = [...cards[0].firstElementChild!.children].map((el) =>
      el.getBoundingClientRect(),
    )
    await expect(name.height).toBe(24)
    await expect(Math.round(prov.top - name.bottom)).toBe(2)
    await expect(Math.round(chips.top - prov.bottom)).toBe(8)
    await expect(chips.height).toBe(20)
  },
}
