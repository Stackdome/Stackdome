import type { ReactNode } from 'react'
import type { Decorator } from '@storybook/react-vite'
import { ReactFlow, ReactFlowProvider, type Edge } from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { ConfirmProvider } from '../src/components/branded/confirm'
import { CurrentUserProvider } from '../src/contexts/current-user-context'
import { StackProvider } from '../src/pages/stacks/contexts/stack-context'
import {
  ReleaseDetailProvider,
  useReleaseDetail,
} from '../src/pages/stacks/components/editor/tabs/deployments/use-release-detail'
import { DEFAULT_PROJECT, ORG_ID, STACK_ID } from './fixtures'
import { ResourceNode, type ResourceFlowNode } from '../src/pages/stacks/components/editor/tabs/architecture/nodes/resource-node'
import { AttachmentNode, type AttachmentFlowNode } from '../src/pages/stacks/components/editor/tabs/architecture/nodes/attachment-node'
import { ConnectionEdge } from '../src/pages/stacks/components/editor/tabs/architecture/edges/connection-edge'

export const withConfirm: Decorator = (Story) => (
  <ConfirmProvider>
    <Story />
  </ConfirmProvider>
)

export const withCurrentUser: Decorator = (Story) => (
  <CurrentUserProvider>
    <Story />
  </CurrentUserProvider>
)

export const withStack: Decorator = (Story) => (
  <StackProvider>
    <Story />
  </StackProvider>
)

export const withHeight = (px: number): Decorator =>
  function HeightDecorator(Story) {
    return (
      <div style={{ height: px }}>
        <Story />
      </div>
    )
  }

// Real useReleaseDetail hook wired to the fixture ids — release-detail fetches
// go through MSW (mock them with releaseHandlers from ./msw-handlers).
function ReleaseDetailHost({ children }: { children: ReactNode }) {
  const value = useReleaseDetail(ORG_ID, DEFAULT_PROJECT, STACK_ID)
  return <ReleaseDetailProvider value={value}>{children}</ReleaseDetailProvider>
}

export const withReleaseDetail: Decorator = (Story) => (
  <ReleaseDetailHost>
    <Story />
  </ReleaseDetailHost>
)

/** Provider-only wrapper for stories whose component uses React Flow hooks. */
export const withFlow: Decorator = (Story) => (
  <ReactFlowProvider>
    <Story />
  </ReactFlowProvider>
)

/** The app's canvas node/edge components, keyed exactly as canvas-editor.tsx does. */
export type FlowNode = ResourceFlowNode | AttachmentFlowNode
const flowNodeTypes = { resource: ResourceNode, attachment: AttachmentNode }
const flowEdgeTypes = { connection: ConnectionEdge }

/**
 * Sized real <ReactFlow> for node/edge stories — Handle needs the flow store,
 * so node components can't render bare. Pass nodes/edges in the app's shapes.
 */
export function FlowHarness({
  nodes,
  edges = [],
  width = 480,
  height = 320,
}: {
  nodes: FlowNode[]
  edges?: Edge[]
  width?: number
  height?: number
}) {
  return (
    <div style={{ width, height }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={flowNodeTypes}
        edgeTypes={flowEdgeTypes}
        fitView
        fitViewOptions={{ maxZoom: 1 }}
        nodesDraggable={false}
        proOptions={{ hideAttribution: true }}
      />
    </div>
  )
}
