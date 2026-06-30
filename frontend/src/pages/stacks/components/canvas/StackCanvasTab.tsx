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

  const dirty = useMemo(
    () => ({
      dirtyResourceIdx: session.dirty.dirtyResourceIdx,
      baselineResourceCount: baselineResources.length,
      pendingDetach: session.pendingDetach,
      baselineAddonIds: connectionAddonIds,
    }),
    [session.dirty, session.pendingDetach, baselineResources.length, connectionAddonIds],
  );

  // Topology + node data (cheap, pure). Re-runs on any edit.
  const dataGraph = useMemo(
    () => deriveGraph({ resources, linkedAddonIds, addonNameById, dirty }),
    [resources, linkedAddonIds, addonNameById, dirty],
  );
  // Signature of the node/edge id-set — changes only when topology changes.
  const topologySignature = useMemo(
    () => `${dataGraph.nodes.map((n) => n.id).join("|")}::${dataGraph.edges.map((e) => e.id).join("|")}`,
    [dataGraph],
  );

  const [nodes, setNodes, onNodesChange] = useNodesState<ResourceFlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [showConnections, setShowConnections] = useState(true);
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

  // Re-layout ONLY when topology changes; preserve in-session drag positions by id.
  useEffect(() => {
    const laid = layoutGraph(dataGraph);
    setNodes((prev) => {
      const posById = new Map(prev.map((n) => [n.id, n.position]));
      return laid.nodes.map((n) => ({ ...n, position: posById.get(n.id) ?? n.position })) as ResourceFlowNode[];
    });
    setEdges(dataGraph.edges as Edge[]);
    // eslint-disable-next-line react-hooks/exhaustive-deps -- intentionally keyed on topology only
  }, [topologySignature, setNodes, setEdges]);

  // Update node data (summary + dirty mark) in place, without moving nodes.
  useEffect(() => {
    const dataById = new Map(dataGraph.nodes.map((n) => [n.id, n.data]));
    setNodes((prev) => prev.map((n) => (dataById.has(n.id) ? { ...n, data: dataById.get(n.id)! } : n)));
  }, [dataGraph, setNodes]);

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
