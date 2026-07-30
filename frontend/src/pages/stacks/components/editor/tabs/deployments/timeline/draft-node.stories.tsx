import type { Meta, StoryObj } from '@storybook/react-vite'
import type { SnapshotDiff } from '../release-snapshot-diff'
import { DraftNode } from './draft-node'

const meta = {
  title: 'Features/Deployments/DraftNode',
  component: DraftNode,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[860px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof DraftNode>

export default meta
type Story = StoryObj<typeof meta>

const stagedDiff: SnapshotDiff = {
  resources: [
    {
      name: 'web',
      change: 'modified',
      sections: [
        {
          kind: 'configuration',
          rows: [
            { key: 'image', from: 'ghcr.io/acme/orders-api:1.4.1', to: 'ghcr.io/acme/orders-api:1.4.2', kind: 'changed' },
          ],
        },
      ],
    },
  ],
  volumes: [],
  connections: [{ name: 'web → orders-db', change: 'added', rows: [] }],
}

// Saved draft awaiting deploy, expanded to show the staged diff vs the live release.
export const Default: Story = {
  args: { phase: 'staged', diff: stagedDiff, vsSeq: 12, isLast: false, defaultOpen: true },
}

// Unsaved edits still syncing — no diff yet, collapsed row with the Unsaved chip.
export const Editing: Story = {
  args: { phase: 'editing', isLast: true },
}
