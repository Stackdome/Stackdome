import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, userEvent, within } from 'storybook/test'
import { StartingPointTabs, type StartingPoint } from './starting-point-tabs'
import { STARTING_POINTS, type Source } from './starting-points'

/** The real five, not a copy of them — a story that restates the product's own
 *  copy is a second place for it to drift. */
const POINTS = STARTING_POINTS

/** The strip is controlled, so a story that never updates `value` would prove
 *  the click handler fires and nothing else. */
function Harness({ initial = 'git' as Source, options = POINTS as StartingPoint<Source>[] }) {
  const [value, setValue] = useState<Source>(initial)
  return (
    <div className="bg-card w-[1186px] max-w-full">
      <StartingPointTabs options={options} value={value} onValueChange={setValue} />
    </div>
  )
}

const meta = {
  title: 'Features/CreateStack/StartingPointTabs',
  component: StartingPointTabs,
  parameters: { layout: 'fullscreen' },
  // Controlled component — every story renders through Harness. These args
  // exist only to satisfy the required props on the type.
  args: { options: POINTS, value: 'git', onValueChange: () => {} },
  render: () => <Harness />,
} satisfies Meta<typeof StartingPointTabs<Source>>

export default meta
type Story = StoryObj<typeof meta>

/** The landing state — the first option live, nothing chosen below it yet.
 *
 *  **There is no label above the strip.** Each title is a verb, so each tab is
 *  complete on its own and there is nothing left to introduce. */
export const Default: Story = {}

/** The blank canvas is the odd one out: its chip is dashed until it is picked. */
export const BlankSelected: Story = {
  render: () => <Harness initial="blank" />,
}

/** The dashed chip fills like every other one when selected — proof that
 *  "selected" reads identically across all five, and that the dashed edge is
 *  identity rather than a state. */
export const DashedChipFillsWhenSelected: Story = {
  render: () => <Harness initial="blank" />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const blank = canvas.getByRole('tab', { name: /Start from nothing/ })
    await expect(blank).toHaveAttribute('aria-selected', 'true')

    const repo = canvas.getByRole('tab', { name: /Deploy your own code/ })
    await expect(repo).toHaveAttribute('aria-selected', 'false')
  },
}

/** Two options only — the strip must not assume five. */
export const TwoOptions: Story = {
  render: () => <Harness options={POINTS.slice(0, 2)} />,
}

/** Long copy has to wrap inside a tab without pushing its siblings out of line. */
export const LongDescription: Story = {
  render: () => (
    <Harness
      options={POINTS.map((p) =>
        p.value === 'template'
          ? { ...p, description: 'n8n, Grafana, Immich, Prometheus, ToolJet, Gitea and rather a lot more besides' }
          : p,
      )}
    />
  ),
}

/** Every title still fits its 204px column on one line — the thing that breaks
 *  when a verb-led title grows. */
export const TitlesFitOnOneLine: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    for (const tab of canvas.getAllByRole('tab')) {
      const title = tab.querySelector('.text-body')!
      await expect(title.getBoundingClientRect().height).toBeLessThanOrEqual(20)
    }
  },
}

/** Arrows move the selection and Home/End jump to the ends — the strip is one
 *  Tab stop, not five. */
export const KeyboardRoving: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const repo = canvas.getByRole('tab', { name: /Deploy your own code/ })
    repo.focus()
    await expect(repo).toHaveAttribute('aria-selected', 'true')

    await userEvent.keyboard('{ArrowRight}')
    await expect(canvas.getByRole('tab', { name: /Run a ready-made app/ })).toHaveAttribute('aria-selected', 'true')

    await userEvent.keyboard('{End}')
    await expect(canvas.getByRole('tab', { name: /Start from nothing/ })).toHaveAttribute('aria-selected', 'true')

    await userEvent.keyboard('{Home}')
    await expect(canvas.getByRole('tab', { name: /Deploy your own code/ })).toHaveAttribute('aria-selected', 'true')
  },
}

/** Clicking swaps the live tab, and only one is ever live. */
export const ClickSelects: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(canvas.getByRole('tab', { name: /Assemble from blocks/ }))

    const tabs = canvas.getAllByRole('tab')
    const live = tabs.filter((t) => t.getAttribute('aria-selected') === 'true')
    await expect(live).toHaveLength(1)
    await expect(live[0]).toHaveAccessibleName(/Assemble from blocks/)
  },
}

/** The group still has an accessible name even though nothing draws it. */
export const GroupIsNamedForScreenReaders: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByRole('tablist')).toHaveAccessibleName('Start from')
  },
}
