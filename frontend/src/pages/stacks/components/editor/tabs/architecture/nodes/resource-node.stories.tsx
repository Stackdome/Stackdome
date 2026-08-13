import type { Meta, StoryObj } from '@storybook/react-vite'
import { NODE_KIND, type ResourceNodeData } from '@/pages/stacks/lib/canvas/graph-from-connections'
import { FlowHarness } from '../../../../../../../../.storybook/decorators'
import type { ResourceFlowNode } from './resource-node'

const meta = {
  title: 'Features/Canvas/ResourceNode',
  tags: ['ai-generated'],
} satisfies Meta

export default meta
type Story = StoryObj

function node(data: Partial<ResourceNodeData>, overrides: Partial<ResourceFlowNode> = {}): ResourceFlowNode {
  return {
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
      ...data,
    },
    ...overrides,
  }
}

export const Web: Story = {
  render: () => (
    <FlowHarness
      nodes={[
        node({
          details: [{ text: 'port 80 · public', port: 80, public: true }],
          portUrls: { 80: 'https://web.example.com' },
        }),
      ]}
    />
  ),
}

export const Postgres: Story = {
  render: () => (
    <FlowHarness
      nodes={[
        node({
          kind: NODE_KIND.addon,
          name: 'postgres',
          kindLabel: 'Postgres',
          glyph: 'postgres',
          brandSlug: 'postgres',
          summary: 'postgres:16',
        }),
      ]}
    />
  ),
}

export const Redis: Story = {
  render: () => (
    <FlowHarness
      nodes={[
        node({
          name: 'cache',
          kindLabel: 'Redis',
          glyph: 'redis',
          brandSlug: 'redis',
          summary: 'redis:7.2',
          details: [{ text: 'port 6379 · internal', port: 6379, public: false }],
        }),
      ]}
    />
  ),
}

export const GitBuildPending: Story = {
  render: () => (
    <FlowHarness
      nodes={[
        node({
          name: 'api',
          kindLabel: 'Service',
          glyph: 'service',
          summary: 'git build',
          dotVariant: 'pending',
        }),
      ]}
    />
  ),
}

export const WithVolumes: Story = {
  render: () => (
    <FlowHarness
      nodes={[
        node({
          volumes: [
            { name: 'web-data', mountPath: '/var/lib/data' },
            { name: 'web-cache', mountPath: '/var/cache' },
          ],
        }),
      ]}
    />
  ),
}

export const Selected: Story = {
  render: () => <FlowHarness nodes={[node({}, { selected: true })]} />,
}

export const LongName: Story = {
  render: () => (
    <FlowHarness
      nodes={[
        node({
          name: 'extremely-long-service-name-that-truncates',
          summary: 'registry.example.com/platform-team/very-long-image-name:2026.07.31-rc1',
        }),
      ]}
    />
  ),
}

export const DirtyEdited: Story = {
  render: () => <FlowHarness nodes={[node({ dirtyState: 'edited' })]} />,
}

export const Removed: Story = {
  render: () => <FlowHarness nodes={[node({ dirtyState: 'removed', dotVariant: 'neutral' })]} />,
}
