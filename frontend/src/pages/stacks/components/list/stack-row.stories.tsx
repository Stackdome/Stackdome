import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { makeStack } from '../../../../../.storybook/fixtures'
import { ReleaseState } from '@/pages/stacks/components/editor/tabs/deployments/release-states'
import type { Stack } from '@/api/stack-types'
import { DeployStackRow, StackRowHeader } from './stack-row'

const meta = {
  title: 'Features/StackRow',
  component: DeployStackRow,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      // The real sheet width at 1440, minus the 12px content edge on each side.
      <div className="w-[1162px]">
        <StackRowHeader />
        <Story />
      </div>
    ),
  ],
  args: { projectName: 'default' },
} satisfies Meta<typeof DeployStackRow>

export default meta
type Story = StoryObj<typeof meta>

const withSource = (over: Partial<Stack>) =>
  makeStack({
    spec: {
      stack_resources: [
        {
          name: 'web',
          source: { git: { repo_url: 'https://github.com/acme/web.git', branch: 'main', commit: 'a3f9d2e1c' } },
        },
      ],
      volumes: [{ name: 'data' }],
    },
    ...over,
  } as Partial<Stack>)

const healthy = withSource({
  latest_release: { id: 'r1', state: ReleaseState.Released },
  converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
})

/**
 * The rest state, and the shape everything else is read against: **one line of
 * status**, 64px tall, with the rule living *inside* the 64 rather than under
 * it.
 */
export const Healthy: Story = {
  args: { stack: healthy, onDelete: fn() },
  play: async ({ canvas }) => {
    const row = canvas.getByRole('link')
    await expect(row.getBoundingClientRect().height).toBe(64)
    await expect(canvas.getByText('Healthy')).toHaveAttribute('data-status-variant', 'ready')
    // The row takes the same glyph as the card — a status that changes shape
    // between the two views costs a re-read every time you switch.
    await expect(canvas.getByText('Healthy').querySelector('svg')).not.toBeNull()
  },
}

/**
 * **The mechanic.** A broken stack grows a second line, so trouble is found by
 * the SHAPE of the list before a word is read or a colour registers — which is
 * also what makes it work for anyone who cannot see the red.
 */
export const FailedWithReason: Story = {
  args: {
    stack: withSource({
      id: 's-docs',
      name: 'docs-site',
      latest_release: {
        id: 'r9',
        state: ReleaseState.Failed,
        message: 'web · image pull failed — ghcr.io/acme/docs:e91a02 not found',
        completed_at: '2026-08-05T11:40:00Z',
      },
    }),
    onDelete: fn(),
  },
  play: async ({ canvas }) => {
    await expect(canvas.getByText('Failed')).toHaveAttribute('data-status-variant', 'error')
    await expect(
      canvas.getByText('web · image pull failed — ghcr.io/acme/docs:e91a02 not found'),
    ).toBeVisible()
    // The row is still 64: the second line grows inside the row, it does not
    // change the list's pitch.
    await expect(canvas.getByRole('link').getBoundingClientRect().height).toBe(64)
  },
}

/**
 * **Deploying stays one line.** In flight is not in trouble, so it gets no
 * reason — and it must not, or the shape stops meaning anything: a poll that
 * turns half the list two-line every deploy is a list that reports nothing.
 *
 * It also shows the word with **no icon**, which is a deliberate call rather
 * than an omission — none of the three families fit an in-flight state.
 */
export const DeployingStaysOneLine: Story = {
  args: {
    stack: withSource({
      id: 's-billing',
      name: 'billing-worker',
      latest_release: { id: 'r2', state: ReleaseState.InProgress, message: 'rolling out 1 of 3' },
      converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
    }),
    onDelete: fn(),
  },
  play: async ({ canvas }) => {
    await expect(canvas.getByText('Deploying')).toBeVisible()
    // The release carries a message, and the row still refuses to show it —
    // the gate is `needsAttention`, not "did the backend write something".
    await expect(canvas.queryByText('rolling out 1 of 3')).toBeNull()
  },
}

/** A long name truncates. It never wraps, and it never widens its column —
 *  the cap is what keeps the status word in the same place on every line. */
export const LongName: Story = {
  args: {
    stack: withSource({
      id: 's-long',
      name: 'internal-platform-notification-dispatch-service-europe-west',
      latest_release: { id: 'r1', state: ReleaseState.Released },
      converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
    }),
    onDelete: fn(),
  },
  play: async ({ canvas }) => {
    const name = canvas.getByText('internal-platform-notification-dispatch-service-europe-west')
    const style = getComputedStyle(name)
    await expect(style.textOverflow).toBe('ellipsis')
    await expect(style.whiteSpace).toBe('nowrap')
  },
}

/** Never deployed is a fact, not a failure — so it colours neutral and carries
 *  no reason line, exactly like a healthy row. */
export const NeverDeployed: Story = {
  args: { stack: makeStack({ id: 's-sandbox', name: 'staging-sandbox' }), onDelete: fn() },
  play: async ({ canvas }) => {
    await expect(canvas.getByText('Not deployed')).toHaveAttribute('data-status-variant', 'neutral')
  },
}

/**
 * **Hovering a row may move nothing.**
 *
 * The board taught this the hard way: hiding the kebab removed it from the
 * auto-layout, the *filling* status column ate its 32px plus the 20px gap, and
 * the status text jumped 52px left on hover. The fix is a permanently reserved
 * 32px grid track and an opacity change — never `display`, never `visibility`.
 *
 * Asserted by rendering the row **with and without** a delete action: if the
 * track were conditional, the status cell would sit in two different places.
 */
export const HoverDoesNotShift: Story = {
  args: { stack: healthy, onDelete: fn() },
  render: (args) => (
    <>
      <DeployStackRow {...args} />
      <DeployStackRow {...args} stack={{ ...args.stack, id: 's-noaction' } as Stack} onDelete={undefined} />
    </>
  ),
  play: async ({ canvas }) => {
    const [withAction, withoutAction] = canvas.getAllByRole('link')
    const statusX = (row: HTMLElement) =>
      Math.round(row.querySelector('[data-slot="status-text"]')!.getBoundingClientRect().left)
    await expect(statusX(withAction)).toBe(statusX(withoutAction))

    // And the hidden control is hidden by OPACITY — it still occupies its track
    // and it still holds its tab stop.
    const kebab = canvas.getByLabelText('Actions for orders-api')
    await expect(parseFloat(getComputedStyle(kebab).opacity)).toBe(0)
    await expect(getComputedStyle(kebab).display).not.toBe('none')
    await expect(kebab.getBoundingClientRect().width).toBeGreaterThan(0)
  },
}
