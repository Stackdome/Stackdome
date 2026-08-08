import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Panel,
  type Edge,
  type OnNodesChange,
  type OnEdgesChange,
  type NodeMouseHandler,
  type OnNodeDrag,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { useCallback, useState } from "react";
import { Lock, Move } from "lucide-react";
import { useTheme } from "@/hooks/use-theme";
import { Popover, PopoverAnchor, PopoverContent } from "@/components/ui/popover";
import { ResourceNode, type ResourceFlowNode } from "./nodes/resource-node";
import { AttachmentNode, type AttachmentFlowNode } from "./nodes/attachment-node";
import { ConnectionEdge } from "./edges/connection-edge";
import { CanvasControls } from "./canvas-controls";
import { AddResourcePopover, AddResourcePanel } from "./add-resource-popover";
import { FIT_OPTIONS } from "./fit-options";

/** Workload nodes (service/addon) plus the compact attachment nodes (secret/volume/object store). */
export type CanvasFlowNode = ResourceFlowNode | AttachmentFlowNode;

/** Declared at module scope — a fresh object identity here re-renders every node. */
const nodeTypes = { resource: ResourceNode, attachment: AttachmentNode };
const edgeTypes = { connection: ConnectionEdge };

/** Edges render as connection edges; styling comes from each edge's data. */
const DEFAULT_EDGE_OPTIONS = { type: "connection" };

/** Snap dragged nodes to a small grid for tidy, lower-frequency position updates. */
const SNAP_GRID: [number, number] = [16, 16];

interface CanvasEditorProps {
  nodes: CanvasFlowNode[];
  edges: Edge[];
  onNodesChange: OnNodesChange<CanvasFlowNode>;
  onEdgesChange: OnEdgesChange;
  onNodeClick?: NodeMouseHandler<CanvasFlowNode>;
  onNodeContextMenu?: NodeMouseHandler<CanvasFlowNode>;
  onNodeDragStart?: OnNodeDrag<CanvasFlowNode>;
  onNodeDrag?: OnNodeDrag<CanvasFlowNode>;
  onNodeDragStop?: OnNodeDrag<CanvasFlowNode>;
  showConnections: boolean;
  onToggleConnections: () => void;
  onAutoLayout: () => void;
  addedBlockIds: string[];
  onAddBlock: (blockId: string) => void;
  addons: { id: string; name: string }[];
  linkedAddonIds: ReadonlySet<string>;
  onLinkAddon: (addonId: string) => void;
  canAddVolume: boolean;
  onAddVolume: () => void;
  /** Live view: nodes are locked in place and every mutation affordance
   *  (add-resource panel, pane context menu) is hidden. */
  readOnly?: boolean;
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
  onNodeContextMenu,
  onNodeDragStart,
  onNodeDrag,
  onNodeDragStop,
  showConnections,
  onToggleConnections,
  onAutoLayout,
  addedBlockIds,
  onAddBlock,
  addons,
  linkedAddonIds,
  onLinkAddon,
  canAddVolume,
  onAddVolume,
  readOnly = false,
}: CanvasEditorProps) {
  // Right-clicking empty canvas opens the same add-resource picker at the
  // cursor (anchored via an invisible fixed-position point).
  const [paneMenuAt, setPaneMenuAt] = useState<{ x: number; y: number } | null>(null);
  const { theme } = useTheme();
  const onPaneContextMenu = useCallback((event: React.MouseEvent | MouseEvent) => {
    event.preventDefault();
    setPaneMenuAt({ x: event.clientX, y: event.clientY });
  }, []);

  return (
    <div className="relative h-full w-full" data-testid="stack-canvas">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        defaultEdgeOptions={DEFAULT_EDGE_OPTIONS}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onNodeContextMenu={onNodeContextMenu}
        onNodeDragStart={onNodeDragStart}
        onNodeDrag={onNodeDrag}
        onNodeDragStop={onNodeDragStop}
        onPaneContextMenu={readOnly ? undefined : onPaneContextMenu}
        nodesDraggable={!readOnly}
        fitView
        fitViewOptions={FIT_OPTIONS}
        // Follow the app's theme toggle, not the OS preference — "system"
        // left the canvas dark while the rest of the UI switched to light.
        colorMode={theme}
        snapToGrid
        snapGrid={SNAP_GRID}
        proOptions={{ hideAttribution: true }}
      >
        {/* bgColor: the Background SVG paints xyflow's default (#141414 in
            dark) over the pane — pin it to the app token so the canvas
            matches the sidebar/chrome in both modes. Two dot layers give the
            grid a fine/bold tier from the dedicated canvas tokens instead of
            xyflow's default dot color. */}
        <Background variant={BackgroundVariant.Dots} gap={24} size={1} color="var(--grid)" bgColor="var(--background)" />
        <Background variant={BackgroundVariant.Dots} gap={120} size={1.5} color="var(--grid-bold)" />
        <CanvasControls
          showConnections={showConnections}
          onToggleConnections={onToggleConnections}
          onAutoLayout={onAutoLayout}
        />
        {!readOnly && (
          <Panel position="top-right">
            <AddResourcePopover
              addedIds={addedBlockIds}
              onAdd={onAddBlock}
              addons={addons}
              linkedAddonIds={linkedAddonIds}
              onLinkAddon={onLinkAddon}
              canAddVolume={canAddVolume}
              onAddVolume={onAddVolume}
            />
          </Panel>
        )}
        {nodes.length > 0 && (
          <Panel position="bottom-center" className="pointer-events-none !mb-[18px]">
            <div className="flex items-center gap-2 text-label text-fg-muted">
              {readOnly ? (
                <>
                  <Lock className="size-[13px]" aria-hidden />
                  viewing the live deployment · read-only · click a card to inspect
                </>
              ) : (
                <>
                  <Move className="size-[13px]" aria-hidden />
                  drag to rearrange · click a card to configure · lines show stack connections
                </>
              )}
            </div>
          </Panel>
        )}
      </ReactFlow>
      {paneMenuAt && (
        <Popover open onOpenChange={(o) => !o && setPaneMenuAt(null)}>
          <PopoverAnchor asChild>
            <span
              aria-hidden
              style={{ position: "fixed", left: paneMenuAt.x, top: paneMenuAt.y, width: 0, height: 0 }}
            />
          </PopoverAnchor>
          <PopoverContent align="start" side="bottom" className="w-[560px] p-0">
            <AddResourcePanel
              addedIds={addedBlockIds}
              onAdd={onAddBlock}
              addons={addons}
              linkedAddonIds={linkedAddonIds}
              onLinkAddon={onLinkAddon}
              canAddVolume={canAddVolume}
              onAddVolume={onAddVolume}
              onRequestClose={() => setPaneMenuAt(null)}
            />
          </PopoverContent>
        </Popover>
      )}
      {nodes.length === 0 && (
        <div className="pointer-events-none absolute inset-0 z-10 flex flex-col items-center justify-center text-center">
          <p className="text-body font-medium text-foreground">No resources yet</p>
          <p className="mt-1 text-body text-muted-foreground">
            Use <span className="font-medium text-foreground">+ Add resource</span> to start building your stack.
          </p>
        </div>
      )}
    </div>
  );
}
