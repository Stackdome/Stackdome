import type { Meta, StoryObj } from '@storybook/react-vite'
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
