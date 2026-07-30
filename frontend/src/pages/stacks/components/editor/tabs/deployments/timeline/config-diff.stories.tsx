import type { Meta, StoryObj } from '@storybook/react-vite'
import type { SnapshotDiff } from '../release-snapshot-diff'
import { ConfigDiff } from './config-diff'

const meta = {
  title: 'Features/Deployments/ConfigDiff',
  component: ConfigDiff,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[720px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ConfigDiff>

export default meta
type Story = StoryObj<typeof meta>

const mixedDiff: SnapshotDiff = {
  resources: [
    {
      name: 'web',
      change: 'modified',
      sections: [
        {
          kind: 'configuration',
          rows: [
            { key: 'image', from: 'ghcr.io/acme/orders-api:1.4.1', to: 'ghcr.io/acme/orders-api:1.4.2', kind: 'changed' },
            { key: 'ports', from: '8080', to: '8080, 9090', kind: 'changed' },
          ],
        },
        {
          kind: 'environment',
          rows: [
            { key: 'LOG_LEVEL', to: 'debug', kind: 'added' },
            { key: 'LEGACY_MODE', from: 'true', kind: 'removed' },
          ],
        },
      ],
    },
    {
      name: 'metrics',
      change: 'added',
      sections: [
        {
          kind: 'configuration',
          rows: [{ key: 'image', to: 'prom/statsd-exporter:v0.26', kind: 'added' }],
        },
      ],
    },
    { name: 'legacy-cron', change: 'removed', sections: [], note: 'Resource removed from the stack.' },
    {
      name: 'worker',
      fromName: 'bg-worker',
      change: 'renamed',
      sections: [],
    },
  ],
  volumes: [
    {
      name: 'uploads',
      change: 'modified',
      rows: [{ key: 'size', from: '5Gi', to: '10Gi', kind: 'changed' }],
    },
  ],
  connections: [
    { name: 'web → orders-db', change: 'added', rows: [] },
  ],
}

export const AddedRemovedChanged: Story = {
  args: { diff: mixedDiff, hasPrev: true, prevSeq: 11 },
}

export const Empty: Story = {
  args: {
    diff: { resources: [], volumes: [], connections: [] },
    hasPrev: true,
    prevSeq: 11,
  },
}

const largeDiff: SnapshotDiff = {
  resources: Array.from({ length: 6 }, (_, i) => ({
    name: `service-${i + 1}`,
    change: 'modified' as const,
    sections: [
      {
        kind: 'environment' as const,
        rows: Array.from({ length: 8 }, (_, j) => ({
          key: `FEATURE_FLAG_${j + 1}`,
          from: 'off',
          to: 'on',
          kind: 'changed' as const,
        })),
      },
    ],
  })),
  volumes: [],
  connections: [],
}

export const LargeDiff: Story = {
  args: { diff: largeDiff, hasPrev: true, prevSeq: 41 },
}
