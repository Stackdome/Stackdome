import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { ViewToggle, useViewMode, type ViewMode } from './view-toggle'

function Harness() {
  const [value, setValue] = useState<ViewMode>('list')
  return <ViewToggle value={value} onValueChange={setValue} />
}

/** Exercises the real hook, so the persistence contract is tested rather than
 *  described. */
function PersistedHarness({ page }: { page: string }) {
  const [value, setValue] = useViewMode(page)
  return (
    <div className="flex items-center gap-3">
      <ViewToggle value={value} onValueChange={setValue} />
      <span data-testid={`mode-${page}`}>{value}</span>
    </div>
  )
}

const meta = {
  title: 'Branded/ViewToggle',
  component: ViewToggle,
  args: { value: 'list', onValueChange: () => {} },
  render: () => <Harness />,
  beforeEach: () => {
    localStorage.removeItem('stackdome.view.demo')
    localStorage.removeItem('stackdome.view.other')
  },
} satisfies Meta<typeof ViewToggle>

export default meta
type Story = StoryObj<typeof meta>

export const List: Story = {}
export const Cards: Story = {
  render: () => {
    const [value, setValue] = useState<ViewMode>('cards')
    return <ViewToggle value={value} onValueChange={setValue} />
  },
}

/** §7 — icon-only, two options, and **List is the default**: it is the denser
 *  view and the one that compares. */
export const ListIsTheDefault: Story = {
  render: () => <PersistedHarness page="demo" />,
  play: async ({ canvas }) => {
    await expect(canvas.getByTestId('mode-demo')).toHaveTextContent('list')
    await expect(canvas.getByRole('radio', { name: 'List' })).toHaveAttribute('aria-checked', 'true')
  },
}

/** Per page, per user. Switching pages does not reset the choice, and two
 *  pages do not share one. */
export const PersistsPerPage: Story = {
  render: () => (
    <div className="flex flex-col gap-3">
      <PersistedHarness page="demo" />
      <PersistedHarness page="other" />
    </div>
  ),
  play: async ({ canvas, userEvent }) => {
    const [demoCards] = canvas.getAllByRole('radio', { name: 'Cards' })
    await userEvent.click(demoCards)

    await expect(canvas.getByTestId('mode-demo')).toHaveTextContent('cards')
    // The second page is untouched — the key is per page, not global.
    await expect(canvas.getByTestId('mode-other')).toHaveTextContent('list')
    await expect(localStorage.getItem('stackdome.view.demo')).toBe('cards')
    await expect(localStorage.getItem('stackdome.view.other')).toBeNull()
  },
}
