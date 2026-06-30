import { useCallback, useEffect, useMemo, useState } from "react";
import { ReactFlowProvider, useNodesState, useEdgesState, type Edge } from "@xyflow/react";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";
import { deriveGraph } from "@/pages/stacks/lib/canvas/derive-graph";
import { layoutGraph } from "@/pages/stacks/lib/canvas/layout-graph";
import { CanvasEditor } from "./CanvasEditor";
import type { ResourceFlowNode } from "./nodes/ResourceNode";

interface StackCanvasTabProps {
  resources: Partial<FormStackResourceData>[];
  linkedAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
}

function StackCanvasFlow({ resources, linkedAddonIds, addonNameById }: StackCanvasTabProps) {
  // Derivation + layout are pure; memoised on their inputs so unrelated renders
  // never reshuffle the board.
  const graph = useMemo(
    () => layoutGraph(deriveGraph({ resources, linkedAddonIds, addonNameById })),
    [resources, linkedAddonIds, addonNameById],
  );

  const [nodes, setNodes, onNodesChange] = useNodesState<ResourceFlowNode>(graph.nodes as ResourceFlowNode[]);
  const [edges, setEdges, onEdgesChange] = useEdgesState(graph.edges as Edge[]);
  const [showConnections, setShowConnections] = useState(true);

  // Re-sync only when topology changes. `graph` is memoised on its inputs, so
  // this never fires on drag (drag mutates `nodes` locally, not `graph`).
  useEffect(() => {
    setNodes(graph.nodes as ResourceFlowNode[]);
    setEdges(graph.edges as Edge[]);
  }, [graph, setNodes, setEdges]);

  const toggleConnections = useCallback(() => setShowConnections((v) => !v), []);

  return (
    <CanvasEditor
      nodes={nodes}
      edges={showConnections ? edges : []}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      showConnections={showConnections}
      onToggleConnections={toggleConnections}
    />
  );
}

/**
 * Flag-gated entry mounted in the Configuration tab. Reads the resolved view
 * inputs (baseline when the edit session is idle, draft when active) and renders
 * the stack as a node graph. The drawer + editing arrive in later slices.
 */
export function StackCanvasTab(props: StackCanvasTabProps) {
  return (
    <div className="h-[calc(100vh-13rem)] overflow-hidden rounded-md border border-border">
      <ReactFlowProvider>
        <StackCanvasFlow {...props} />
      </ReactFlowProvider>
    </div>
  );
}
