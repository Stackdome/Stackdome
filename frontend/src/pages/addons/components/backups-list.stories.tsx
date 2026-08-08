import type { Meta, StoryObj } from '@storybook/react-vite'
import { makeBackup } from '../../../../.storybook/fixtures'
import { BackupsList } from './backups-list'

const meta = {
  title: 'Features/Addons/BackupsList',
  component: BackupsList,
  tags: ['ai-generated'],
} satisfies Meta<typeof BackupsList>

export default meta
type Story = StoryObj<typeof meta>

export const WithBackups: Story = {
  args: {
    backups: [
      makeBackup(),
      // ScheduledBackup-spawned run reported as "manual": the name pattern
      // drives the displayed "Scheduled" type.
      makeBackup({
        id: 'bk-2',
        name: 'orders-db-scheduled-backup-20260728020000',
        started_at: '2026-07-28T02:00:00Z',
        completed_at: '2026-07-28T02:03:10Z',
        size_bytes: 698351616,
      }),
      makeBackup({
        id: 'bk-3',
        name: 'orders-db-backup-3',
        phase: 'failed',
        error: 'wal-archive: connection to object store timed out',
        completed_at: undefined,
        size_bytes: undefined,
      }),
    ],
  },
}

export const Empty: Story = {
  args: { backups: [] },
}

export const InProgress: Story = {
  args: {
    backups: [
      makeBackup({
        id: 'bk-4',
        name: 'orders-db-backup-4',
        phase: 'running',
        completed_at: undefined,
        size_bytes: undefined,
      }),
      makeBackup({
        id: 'bk-5',
        name: 'orders-db-backup-5',
        phase: 'pending',
        started_at: undefined,
        completed_at: undefined,
        size_bytes: undefined,
      }),
    ],
  },
}
