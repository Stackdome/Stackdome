import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Panel,
  type Edge,
  type OnNodesChange,
  type OnEdgesChange,
  type NodeMouseHandler,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { Move } from "lucide-react";
import { ResourceNode, type ResourceFlowNode } from "./nodes/ResourceNode";
import { ConnectionEdge } from "./edges/ConnectionEdge";
import { CanvasControls } from "./CanvasControls";
import { AddResourcePopover } from "./AddResourcePopover";
import { FIT_OPTIONS } from "./fit-options";

/** Declared at module scope — a fresh object identity here re-renders every node. */
const nodeTypes = { resource: ResourceNode };
const edgeTypes = { connection: ConnectionEdge };

/** All derived edges are variable-reference connections (dashed amber). */
const DEFAULT_EDGE_OPTIONS = { type: "connection" };

/** Snap dragged nodes to a small grid for tidy, lower-frequency position updates. */
const SNAP_GRID: [number, number] = [16, 16];

interface CanvasEditorProps {
  nodes: ResourceFlowNode[];
  edges: Edge[];
  onNodesChange: OnNodesChange<ResourceFlowNode>;
  onEdgesChange: OnEdgesChange;
  onNodeClick?: NodeMouseHandler<ResourceFlowNode>;
  showConnections: boolean;
  onToggleConnections: () => void;
  onAutoLayout: () => void;
  addedBlockIds: string[];
  onAddBlock: (blockId: string) => void;
  addons: { id: string; name: string }[];
  linkedAddonIds: ReadonlySet<string>;
  onLinkAddon: (addonId: string) => void;
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
  onAutoLayout,
  addedBlockIds,
  onAddBlock,
  addons,
  linkedAddonIds,
  onLinkAddon,
}: CanvasEditorProps) {
  return (
    <div className="h-full w-full" data-testid="stack-canvas">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        defaultEdgeOptions={DEFAULT_EDGE_OPTIONS}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        fitView
        fitViewOptions={FIT_OPTIONS}
        colorMode="system"
        snapToGrid
        snapGrid={SNAP_GRID}
        proOptions={{ hideAttribution: true }}
      >
        <Background variant={BackgroundVariant.Dots} gap={24} size={1} />
        {nodes.length === 0 && (
          <Panel position="top-center">
            <div className="mt-24 text-center">
              <p className="text-sm font-medium text-foreground">No resources yet</p>
              <p className="mt-1 text-[13px] text-muted-foreground">
                Use <span className="font-medium text-foreground">+ Add resource</span> to start building your stack.
              </p>
            </div>
          </Panel>
        )}
        <CanvasControls
          showConnections={showConnections}
          onToggleConnections={onToggleConnections}
          onAutoLayout={onAutoLayout}
        />
        <Panel position="top-left">
          <AddResourcePopover
            addedIds={addedBlockIds}
            onAdd={onAddBlock}
            addons={addons}
            linkedAddonIds={linkedAddonIds}
            onLinkAddon={onLinkAddon}
          />
        </Panel>
        {nodes.length > 0 && (
          <Panel position="bottom-center" className="pointer-events-none !mb-[18px]">
            <div className="flex items-center gap-2 text-[11.5px] text-fg-muted">
              <Move className="size-[13px]" aria-hidden />
              drag to rearrange · click a node to configure · edges carry connection env vars
            </div>
          </Panel>
        )}
      </ReactFlow>
    </div>
  );
}
