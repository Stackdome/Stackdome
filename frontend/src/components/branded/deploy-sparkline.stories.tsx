import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { DeploySparkline } from './deploy-sparkline'

const meta = {
  title: 'Branded/DeploySparkline',
  component: DeploySparkline,
  decorators: [(Story) => <div className="w-64">{Story()}</div>],
} satisfies Meta<typeof DeploySparkline>

export default meta
type Story = StoryObj<typeof meta>

const BUSY = [3, 1, 0, 4, 2, 0, 0, 6, 1, 2, 0, 3, 5, 1]
const QUIET = [0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1]

export const Busy: Story = { args: { history: BUSY } }
export const Quiet: Story = { args: { history: QUIET } }
export const OneBigDay: Story = { args: { history: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 19] } }

/** A stack that has never deployed says so in words. Fourteen invented bars for
 *  a stack that never shipped is the failure this component exists to avoid. */
export const NeverDeployed: Story = {
  args: { history: undefined },
  play: async ({ canvas }) => {
    await expect(canvas.getByText('No deploys yet')).toBeInTheDocument()
    await expect(canvas.queryByRole('img')).toBeNull()
  },
}

/** Deployed once, then went quiet. This is NOT the same as never deployed, and
 *  the two must not render the same way — a flat line is a real statement. */
export const DeployedThenWentQuiet: Story = {
  args: { history: [2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0] },
  play: async ({ canvas }) => {
    await expect(canvas.queryByText('No deploys yet')).toBeNull()
    await expect(canvas.getByRole('img')).toBeInTheDocument()
  },
}

/** An all-zero fortnight must not divide by zero, and must not vanish. */
export const AllZero: Story = {
  args: { history: Array(14).fill(0) },
  play: async ({ canvas }) => {
    const chart = canvas.getByRole('img')
    await expect(chart).toHaveAttribute('aria-label', '0 deploys in the last 14 days')
    await expect(chart.children).toHaveLength(14)
  },
}

/** The bars report VOLUME, not health. Colouring a fortnight of history with
 *  today's status would repaint a good fortnight red the moment a deploy fails
 *  — the worst error in the first draft of the card. */
export const BarsAreNeutral: Story = {
  args: { history: BUSY },
  play: async ({ canvas }) => {
    const bars = [...canvas.getByRole('img').children] as HTMLElement[]

    const probe = document.createElement('div')
    document.body.appendChild(probe)
    const token = (cls: string) => {
      probe.className = cls
      return getComputedStyle(probe).backgroundColor
    }
    const [fg2, danger, success, warn] = ['bg-fg-2', 'bg-danger', 'bg-success', 'bg-warn'].map(token)
    probe.remove()

    for (const bar of bars.filter((b) => b.style.height !== '1px')) {
      const bg = getComputedStyle(bar).backgroundColor
      await expect(bg).toBe(fg2)
      for (const status of [danger, success, warn]) {
        await expect(bg).not.toBe(status)
      }
    }
  },
}

/** It says what it counts. A deploy count and a CPU trace look identical at
 *  this size, so the label is not decoration. */
export const Labelled: Story = {
  args: { history: BUSY },
  play: async ({ canvas }) => {
    await expect(canvas.getByText('Deploys · 14d')).toBeInTheDocument()
    await expect(canvas.getByRole('img')).toHaveAttribute(
      'aria-label',
      '28 deploys in the last 14 days',
    )
  },
}

/** A quiet day keeps a 1px stub rather than a gap — the fortnight has to read
 *  as one continuous span, or a quiet week looks like missing data. */
export const QuietDaysKeepTheirPlace: Story = {
  args: { history: QUIET },
  play: async ({ canvas }) => {
    const bars = [...canvas.getByRole('img').children] as HTMLElement[]
    await expect(bars).toHaveLength(14)
    const stubs = bars.filter((b) => b.style.height === '1px')
    await expect(stubs).toHaveLength(12)
    for (const stub of stubs) {
      await expect(stub.getBoundingClientRect().width).toBeGreaterThan(0)
    }
  },
}
