import type { Meta, StoryObj } from '@storybook/react-vite'
import {
  ATTACHMENT_LABEL,
  NODE_KIND,
  type AttachmentKind,
} from '@/pages/stacks/lib/canvas/graph-from-connections'
import { FlowHarness } from '../../../../../../../../.storybook/decorators'
import type { AttachmentFlowNode } from './attachment-node'

const meta = {
  title: 'Features/Canvas/AttachmentNode',
  tags: ['ai-generated'],
} satisfies Meta

export default meta
type Story = StoryObj

function node(kind: AttachmentKind, name: string): AttachmentFlowNode {
  return {
    id: `${kind}:${name}`,
    type: 'attachment',
    position: { x: 0, y: 0 },
    data: { kind, name, kindLabel: ATTACHMENT_LABEL[kind] },
  }
}

export const Secret: Story = {
  render: () => <FlowHarness nodes={[node(NODE_KIND.secret, 'api-credentials')]} />,
}

export const Volume: Story = {
  render: () => <FlowHarness nodes={[node(NODE_KIND.volume, 'web-data')]} />,
}

export const ObjectStore: Story = {
  render: () => <FlowHarness nodes={[node(NODE_KIND.objectStore, 'uploads-bucket')]} />,
}

export const LongName: Story = {
  render: () => (
    <FlowHarness nodes={[node(NODE_KIND.secret, 'extremely-long-secret-name-that-truncates-in-the-card')]} />
  ),
}
