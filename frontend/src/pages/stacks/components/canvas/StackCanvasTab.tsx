import { useCallback, useEffect, useMemo, useState } from "react";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import {
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  useReactFlow,
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
import { addBlockToStack } from "@/pages/stacks/lib/block-to-form";
import { blockCatalog, getBlockById } from "@/pages/stacks/data/blocks/registry";
import { CanvasEditor } from "./CanvasEditor";
import { ResourceDrawer } from "./ResourceDrawer";
import { VolumeDrawer } from "./VolumeDrawer";
import { DrawerStack, type DrawerPanelDescriptor } from "./DrawerStack";
import {
  replaceStack,
  pushEntry,
  truncateTo,
  popEntry,
  entryKey,
  type DrawerEntry,
} from "@/pages/stacks/lib/canvas/drawer-stack";
import { nodePresentation } from "@/pages/stacks/lib/canvas/node-presentation";
import { NodeGlyph } from "./nodes/node-glyph";
import { HardDrive } from "lucide-react";
import { FIT_OPTIONS } from "./fit-options";
import type { ResourceFlowNode } from "./nodes/ResourceNode";

interface StackCanvasTabProps {
  session: UseStackEditSession;
  /** Diff baseline (the deployed release snapshot when one exists). */
  baselineResources: Partial<FormStackResourceData>[];
  baselineVolumes: Partial<VolumeFormData>[];
  /** Current server state — what the canvas shows when no session is active,
   *  and what a lazily-started session's working draft seeds from. */
  draftResources: Partial<FormStackResourceData>[];
  draftVolumes: Partial<VolumeFormData>[];
  connectionAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
  errors: { [index: number]: { [field: string]: string | undefined } };
  /** Switch the editor to the Logs tab (from the drawer's "View logs"). */
  onViewLogs?: () => void;
}

function StackCanvasFlow({
  session,
  baselineResources,
  baselineVolumes,
  draftResources,
  draftVolumes,
  connectionAddonIds,
  addonNameById,
  errors,
  onViewLogs,
}: StackCanvasTabProps) {
  // Read from the live draft when the session is active, server state otherwise.
  const resources = session.isActive ? session.draft.resources : draftResources;
  const linkedAddonIds = session.isActive ? session.linkedAddonIds : connectionAddonIds;

  const dirty = useMemo(
    () => ({
      dirtyResourceIdx: session.dirty.dirtyResourceIdx,
      baselineResourceCount: baselineResources.length,
      baselineAddonIds: connectionAddonIds,
    }),
    [session.dirty, baselineResources.length, connectionAddonIds],
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
  const [drawerStack, setDrawerStack] = useState<DrawerEntry[]>([]);
  const { fitView } = useReactFlow();

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

  // Re-run auto-layout: reset every node to its fresh dagre position and re-fit.
  const autoLayout = useCallback(() => {
    const laid = layoutGraph(dataGraph);
    setNodes(laid.nodes as ResourceFlowNode[]);
    requestAnimationFrame(() => fitView(FIT_OPTIONS));
  }, [dataGraph, setNodes, fitView]);

  const onNodeClick = useCallback<NodeMouseHandler<ResourceFlowNode>>(
    (_event, node) => {
      const idx = node.data.resourceIdx;
      if (idx == null) return; // addon node — managed via the Environment tab, no drawer in v1
      // Activate an edit session lazily so drawer edits land in a draft.
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          {
            linkedAddonIds: new Set(connectionAddonIds),
            openResourceIdx: idx,
            openTab: "configuration",
            draft: { resources: draftResources, volumes: draftVolumes },
          },
        );
      }
      setDrawerStack(replaceStack({ kind: "resource", index: idx }));
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  // Block ids already present in the stack (drives the picker's "added" badge).
  const addedBlockIds = useMemo(() => {
    const names = new Set(resources.map((r) => r.name));
    return blockCatalog
      .filter((b) => names.has(b.id) || [...names].some((n) => n?.startsWith(`${b.id}-`)))
      .map((b) => b.id);
  }, [resources]);

  const onAddBlock = useCallback(
    (blockId: string) => {
      const block = getBlockById(blockId);
      if (!block) return;
      const current = session.isActive
        ? { resources: session.draft.resources, volumes: session.draft.volumes }
        : { resources: draftResources, volumes: draftVolumes };
      const working = { name: "", labels: [], spec: { stack_resources: current.resources, volumes: current.volumes } };
      // eslint-disable-next-line @typescript-eslint/no-explicit-any -- WorkingStack uses full FormStackResourceData; the draft holds Partial, structurally compatible here
      const next = addBlockToStack(working as any, block);
      // block-to-form yields base FormVolumeData; the session works in the extended
      // form (adds an optional sourceType) — structurally compatible at runtime.
      const nextVolumes = next.spec.volumes as unknown as VolumeFormData[];
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          {
            linkedAddonIds: new Set(connectionAddonIds),
            draft: { resources: next.spec.stack_resources, volumes: nextVolumes },
          },
        );
      } else {
        session.updateResources(() => next.spec.stack_resources);
        session.updateVolumes(() => nextVolumes);
      }
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  const { addons: allAddons } = usePostgresAddons();
  const pickableAddons = useMemo(
    () => allAddons.filter((a) => a.id && a.name).map((a) => ({ id: a.id!, name: a.name! })),
    [allAddons],
  );
  const onLinkAddon = useCallback(
    (addonId: string) => {
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          {
            linkedAddonIds: new Set([...connectionAddonIds, addonId]),
            draft: { resources: draftResources, volumes: draftVolumes },
          },
        );
      } else {
        session.setLinkedAddonIds((prev) => new Set(prev).add(addonId));
      }
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  const popDrawer = useCallback(() => setDrawerStack((s) => popEntry(s)), []);
  const closeAllDrawers = useCallback(() => setDrawerStack([]), []);
  const truncateDrawers = useCallback((depth: number) => setDrawerStack((s) => truncateTo(s, depth)), []);
  const openVolume = useCallback(
    (name: string) => {
      // Guard dangling mount references (mount rows can outlive a deleted
      // volume): pushing one would render an empty panel.
      const volumes = session.isActive ? session.draft.volumes : draftVolumes;
      if (!volumes.some((v) => v.name === name)) return;
      setDrawerStack((s) => pushEntry(s, { kind: "volume", name }));
    },
    [session, draftVolumes],
  );
  const removeResource = useCallback(
    (idx: number) => {
      session.updateResources((prev) => prev.filter((_, i) => i !== idx));
      setDrawerStack([]);
    },
    [session],
  );

  // Drop panels whose target no longer exists in the draft (deleted resource/volume).
  useEffect(() => {
    setDrawerStack((s) => {
      const next = s.filter((e) =>
        e.kind === "resource"
          ? e.index < resources.length
          : (session.isActive ? session.draft.volumes : draftVolumes).some((v) => v.name === e.name),
      );
      // Same-ref bailout: skip the state update (and re-render) when nothing was dropped.
      return next.length === s.length ? s : next;
    });
  }, [resources.length, session.isActive, session.draft.volumes, draftVolumes]);

  const panels: DrawerPanelDescriptor[] = useMemo(
    () =>
      drawerStack.map((entry) => {
        if (entry.kind === "resource") {
          const r = resources[entry.index] ?? {};
          const pres = nodePresentation({
            isAddon: false,
            image: r.image_spec?.image,
            hasBuild: !!r.build_spec,
            ports: (r.ports ?? []).map((p) => ({
              number: p.number,
              protocol: p.protocol,
              exposedToPublic: p.exposed_to_public,
            })),
          });
          return {
            entry,
            title: r.name || `Resource ${entry.index + 1}`,
            icon: <NodeGlyph glyph={pres.glyph} className="size-[19px]" />,
          };
        }
        return { entry, title: entry.name, icon: <HardDrive className="size-[19px]" /> };
      }),
    [drawerStack, resources],
  );

  const frontEntry = drawerStack[drawerStack.length - 1];
  const frontBody =
    frontEntry?.kind === "resource" ? (
      <ResourceDrawer
        key={entryKey(frontEntry)}
        resourceIndex={frontEntry.index}
        session={session}
        baselineResources={baselineResources}
        connectionAddonIds={connectionAddonIds}
        errors={errors[frontEntry.index] ?? {}}
        onClose={popDrawer}
        onRemove={removeResource}
        onViewLogs={onViewLogs}
        onOpenVolume={openVolume}
      />
    ) : frontEntry ? (
      <VolumeDrawer key={entryKey(frontEntry)} volumeName={frontEntry.name} session={session} onClose={popDrawer} />
    ) : null;

  return (
    <div className="flex h-full w-full">
      <div className="relative min-w-0 flex-1">
        <CanvasEditor
          nodes={nodes}
          edges={showConnections ? edges : []}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={onNodeClick}
          showConnections={showConnections}
          onToggleConnections={toggleConnections}
          onAutoLayout={autoLayout}
          addedBlockIds={addedBlockIds}
          onAddBlock={onAddBlock}
          addons={pickableAddons}
          linkedAddonIds={linkedAddonIds}
          onLinkAddon={onLinkAddon}
        />
      </div>
      <DrawerStack
        panels={panels}
        front={frontBody}
        onTruncate={truncateDrawers}
        onPop={popDrawer}
        onCloseAll={closeAllDrawers}
      />
    </div>
  );
}

/**
 * Flag-gated entry mounted in the Configuration tab. Renders the stack as a node
 * graph; clicking a service node opens the config drawer. The `relative` wrapper
 * anchors the drawer overlay.
 */
export function StackCanvasTab(props: StackCanvasTabProps) {
  return (
    // Edge-to-edge inside the full-bleed editor shell (shell owns the chrome).
    <div className="relative h-full w-full overflow-hidden">
      <ReactFlowProvider>
        <StackCanvasFlow {...props} />
      </ReactFlowProvider>
    </div>
  );
}
