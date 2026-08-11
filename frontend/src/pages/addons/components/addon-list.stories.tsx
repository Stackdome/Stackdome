import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { makeAddon } from '../../../../.storybook/fixtures'
import { AddonList } from './addon-list'

const mixed = [
  makeAddon(),
  makeAddon({
    id: 'pg-2',
    name: 'analytics-db',
    status: { state: 'Creating' },
    created_at: '2026-07-30T18:00:00Z',
  }),
  makeAddon({
    id: 'pg-3',
    name: 'sessions-db',
    status: { state: 'Error', message: 'volume provisioning failed' },
    created_at: '2026-07-25T12:00:00Z',
  }),
  makeAddon({
    id: 'pg-4',
    name: 'archive-db',
    status: { state: 'Hibernated' },
    created_at: '2026-06-01T08:00:00Z',
  }),
]

const meta = {
  title: 'Features/Addons/AddonList',
  component: AddonList,
  tags: ['ai-generated'],
} satisfies Meta<typeof AddonList>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: { addons: mixed },
}

export const Empty: Story = {
  args: { addons: [] },
}

export const ReadOnly: Story = {
  args: { addons: mixed, canWrite: () => false },
}

// Rubric D2/D4: the active status filter is an ink tint, never brand orange
// — hue reports real addon state (ready/pending/error), not UI selection.
export const ActiveFilterIsInkTint: Story = {
  args: { addons: mixed },
  play: async ({ canvas, userEvent }) => {
    const readyFilter = canvas.getByRole('button', { name: /ready/i })
    await userEvent.click(readyFilter)
    const style = getComputedStyle(readyFilter)
    const brand = getComputedStyle(document.documentElement).getPropertyValue('--brand').trim()
    const probe = document.createElement('div')
    probe.style.color = brand
    document.body.appendChild(probe)
    const brandColor = getComputedStyle(probe).color
    probe.remove()
    await expect(style.color).not.toBe(brandColor)
  },
}

// Row hover is a fill shift only (D13) — the addon name must not change
// color on hover, and the row itself must not move or scale.
export const RowHoverNoColorOrTransform: Story = {
  args: { addons: mixed },
  play: async ({ canvas, userEvent }) => {
    const link = canvas.getAllByRole('link')[0]
    const name = link.querySelector('span.font-medium') as HTMLElement
    const colorBefore = getComputedStyle(name).color
    const transformBefore = getComputedStyle(link).transform
    await userEvent.hover(link)
    await expect(getComputedStyle(name).color).toBe(colorBefore)
    await expect(getComputedStyle(link).transform).toBe(transformBefore)
  },
}
