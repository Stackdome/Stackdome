import {
  ReactFlow,
  Background,
  Controls,
  Panel,
  type Edge,
  type OnNodesChange,
  type OnEdgesChange,
  type NodeMouseHandler,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { ResourceNode, type ResourceFlowNode } from "./nodes/ResourceNode";
import { AddResourcePopover } from "./AddResourcePopover";

/** Declared at module scope — a fresh object identity here re-renders every node. */
const nodeTypes = { resource: ResourceNode };

interface CanvasEditorProps {
  nodes: ResourceFlowNode[];
  edges: Edge[];
  onNodesChange: OnNodesChange<ResourceFlowNode>;
  onEdgesChange: OnEdgesChange;
  onNodeClick?: NodeMouseHandler<ResourceFlowNode>;
  showConnections: boolean;
  onToggleConnections: () => void;
  addedBlockIds: string[];
  onAddBlock: (blockId: string) => void;
}

/**
 * React Flow render surface. It owns no stack state — nodes/edges are derived
 * upstream and handed in. Keeps the surface a dumb, memo-friendly view.
 */
export function CanvasEditor({
  nodes,
  edges,
  onNodesChange,
  onEdgesChange,
  onNodeClick,
  showConnections,
  onToggleConnections,
  addedBlockIds,
  onAddBlock,
}: CanvasEditorProps) {
  return (
    <div className="h-full w-full" data-testid="stack-canvas">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        fitView
        colorMode="system"
        proOptions={{ hideAttribution: true }}
      >
        <Background />
        <Controls />
        <Panel position="top-left">
          <AddResourcePopover addedIds={addedBlockIds} onAdd={onAddBlock} />
        </Panel>
        <Panel position="top-right">
          <button
            type="button"
            onClick={onToggleConnections}
            className="rounded-md border border-border bg-card px-2 py-1 font-mono text-[11px] uppercase tracking-wider text-fg-muted transition-colors hover:text-foreground"
          >
            {showConnections ? "Hide connections" : "Show connections"}
          </button>
        </Panel>
      </ReactFlow>
    </div>
  );
}
