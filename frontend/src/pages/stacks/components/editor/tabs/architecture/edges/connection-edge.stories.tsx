import type { Meta, StoryObj } from '@storybook/react-vite'
import {
  EDGE_KIND,
  EDGE_SOURCE_OF_TRUTH,
  NODE_KIND,
} from '@/pages/stacks/lib/canvas/graph-from-connections'
import { FlowHarness, type FlowNode } from '../../../../../../../../.storybook/decorators'

const meta = {
  title: 'Features/Canvas/ConnectionEdge',
  tags: ['ai-generated'],
} satisfies Meta

export default meta
type Story = StoryObj

const nodes: FlowNode[] = [
  {
    id: 'resource:web',
    type: 'resource',
    position: { x: 0, y: 0 },
    data: {
      kind: NODE_KIND.service,
      name: 'web',
      kindLabel: 'Web',
      glyph: 'web',
      dotVariant: 'ready',
      summary: 'nginx:1.27',
      volumes: [],
    },
  },
  {
    id: 'addon:postgres',
    type: 'resource',
    position: { x: 340, y: 120 },
    data: {
      kind: NODE_KIND.addon,
      name: 'postgres',
      kindLabel: 'Postgres',
      glyph: 'postgres',
      brandSlug: 'postgres',
      dotVariant: 'ready',
      summary: 'postgres:16',
      volumes: [],
    },
  },
]

export const Default: Story = {
  render: () => (
    <FlowHarness
      nodes={nodes}
      edges={[
        {
          id: 'e1',
          source: 'addon:postgres',
          target: 'resource:web',
          type: 'connection',
          data: { kind: EDGE_KIND.env, sourceOfTruth: EDGE_SOURCE_OF_TRUTH.connection },
        },
      ]}
    />
  ),
}
