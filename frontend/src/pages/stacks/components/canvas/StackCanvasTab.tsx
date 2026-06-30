import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  type Edge,
  type NodeMouseHandler,
} from "@xyflow/react";
import type {
  FormStackResourceData,
  FormVolumeExtendedData as VolumeFormData,
} from "@/pages/stacks/schemas/form-schema";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import { deriveGraph } from "@/pages/stacks/lib/canvas/derive-graph";
import { layoutGraph } from "@/pages/stacks/lib/canvas/layout-graph";
import { CanvasEditor } from "./CanvasEditor";
import { ResourceDrawer } from "./ResourceDrawer";
import type { ResourceFlowNode } from "./nodes/ResourceNode";

interface StackCanvasTabProps {
  session: UseStackEditSession;
  baselineResources: Partial<FormStackResourceData>[];
  baselineVolumes: Partial<VolumeFormData>[];
  connectionAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
  errors: { [index: number]: { [field: string]: string | undefined } };
}

function StackCanvasFlow({
  session,
  baselineResources,
  baselineVolumes,
  connectionAddonIds,
  addonNameById,
  errors,
}: StackCanvasTabProps) {
  // Read from the live draft when the session is active, baseline otherwise.
  const resources = session.isActive ? session.draft.resources : baselineResources;
  const linkedAddonIds = session.isActive ? session.linkedAddonIds : connectionAddonIds;

  const graph = useMemo(
    () => layoutGraph(deriveGraph({ resources, linkedAddonIds, addonNameById })),
    [resources, linkedAddonIds, addonNameById],
  );

  const [nodes, setNodes, onNodesChange] = useNodesState<ResourceFlowNode>(graph.nodes as ResourceFlowNode[]);
  const [edges, setEdges, onEdgesChange] = useEdgesState(graph.edges as Edge[]);
  const [showConnections, setShowConnections] = useState(true);
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

  // Re-sync only on topology change (graph is memoised on its inputs).
  useEffect(() => {
    setNodes(graph.nodes as ResourceFlowNode[]);
    setEdges(graph.edges as Edge[]);
  }, [graph, setNodes, setEdges]);

  const toggleConnections = useCallback(() => setShowConnections((v) => !v), []);

  const onNodeClick = useCallback<NodeMouseHandler<ResourceFlowNode>>(
    (_event, node) => {
      const idx = node.data.resourceIdx;
      if (idx == null) return; // addon node — managed via the Environment tab, no drawer in v1
      // Activate an edit session lazily so drawer edits land in a draft.
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          { linkedAddonIds: new Set(connectionAddonIds), openResourceIdx: idx, openTab: "configuration" },
        );
      }
      setSelectedIndex(idx);
    },
    [session, baselineResources, baselineVolumes, connectionAddonIds],
  );

  const closeDrawer = useCallback(() => setSelectedIndex(null), []);
  const removeResource = useCallback(
    (idx: number) => {
      session.updateResources((prev) => prev.filter((_, i) => i !== idx));
      setSelectedIndex(null);
    },
    [session],
  );

  return (
    <>
      <CanvasEditor
        nodes={nodes}
        edges={showConnections ? edges : []}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        showConnections={showConnections}
        onToggleConnections={toggleConnections}
      />
      {selectedIndex != null && (
        <ResourceDrawer
          key={selectedIndex}
          resourceIndex={selectedIndex}
          session={session}
          baselineResources={baselineResources}
          baselineVolumes={baselineVolumes}
          connectionAddonIds={connectionAddonIds}
          errors={errors[selectedIndex] ?? {}}
          onClose={closeDrawer}
          onRemove={removeResource}
        />
      )}
    </>
  );
}

/**
 * Flag-gated entry mounted in the Configuration tab. Renders the stack as a node
 * graph; clicking a service node opens the config drawer. The `relative` wrapper
 * anchors the drawer overlay.
 */
export function StackCanvasTab(props: StackCanvasTabProps) {
  return (
    <div className="relative h-[calc(100vh-13rem)] overflow-hidden rounded-md border border-border">
      <ReactFlowProvider>
        <StackCanvasFlow {...props} />
      </ReactFlowProvider>
    </div>
  );
}
