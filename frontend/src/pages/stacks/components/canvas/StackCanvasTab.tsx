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
import type { EditSessionDraft, UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import { useSecrets } from "@/pages/stacks/hooks/use-secrets";
import { useStackTopology } from "@/pages/stacks/hooks/use-stack-topology";
import { deriveGraph } from "@/pages/stacks/lib/canvas/graph-from-connections";
import { mergeTopology } from "@/pages/stacks/lib/canvas/merge-topology";
import { layoutGraph } from "@/pages/stacks/lib/canvas/layout-graph";
import { addBlockToStack } from "@/pages/stacks/lib/block-to-form";
import { blockCatalog, getBlockById } from "@/pages/stacks/data/blocks/registry";
import { CanvasEditor } from "./CanvasEditor";
import { ResourceDrawer } from "./ResourceDrawer";
import { VolumeDrawer } from "./VolumeDrawer";
import { AddVolumeDialog } from "./AddVolumeDialog";
import { addMount, newVolume } from "@/pages/stacks/lib/canvas/volume-ops";
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
import type { CanvasFlowNode } from "./CanvasEditor";

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
  /** Null for draft (unsaved) stacks — no server topology exists yet. */
  topologyIds: { orgId: string; teamName: string; stackId: string } | null;
  /** Bump to force a topology refetch (wired to autosave refreshes). */
  topologyRefreshKey: number;
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
  topologyIds,
  topologyRefreshKey,
}: StackCanvasTabProps) {
  // Read from the live draft when the session is active, server state otherwise.
  const resources = session.isActive ? session.draft.resources : draftResources;
  const linkedAddonIds = session.isActive ? session.linkedAddonIds : connectionAddonIds;
  const volumes = session.isActive ? session.draft.volumes : draftVolumes;
  const volumeNames = useMemo(
    () => volumes.map((v) => v.name).filter((n): n is string => !!n),
    [volumes],
  );

  const dirty = useMemo(
    () => ({
      dirtyResourceIdx: session.dirty.dirtyResourceIdx,
      baselineResourceCount: baselineResources.length,
      baselineAddonIds: connectionAddonIds,
    }),
    [session.dirty, baselineResources.length, connectionAddonIds],
  );

  const { secrets } = useSecrets();
  const secretNameById = useMemo(
    () => new Map(secrets.filter((s) => s.id && s.name).map((s) => [s.id!, s.name!])),
    [secrets],
  );

  const { topology } = useStackTopology({ ids: topologyIds, refreshKey: topologyRefreshKey });

  // Local connection-derived data (cheap, pure). Re-runs on any edit.
  const dataGraph = useMemo(
    () => deriveGraph({ resources, linkedAddonIds, addonNameById, secretNameById, volumeNames, dirty }),
    [resources, linkedAddonIds, addonNameById, secretNameById, volumeNames, dirty],
  );
  // Local graph enhanced with server-derived edges + runtime status.
  const mergedGraph = useMemo(() => mergeTopology(dataGraph, topology), [dataGraph, topology]);
  // Signature of the node/edge id-set — changes only when topology changes.
  const topologySignature = useMemo(
    () => `${mergedGraph.nodes.map((n) => n.id).join("|")}::${mergedGraph.edges.map((e) => e.id).join("|")}`,
    [mergedGraph],
  );

  const [nodes, setNodes, onNodesChange] = useNodesState<CanvasFlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [showConnections, setShowConnections] = useState(true);
  const [drawerStack, setDrawerStack] = useState<DrawerEntry[]>([]);
  const { fitView } = useReactFlow();

  // Re-layout ONLY when topology changes; preserve in-session drag positions by id.
  useEffect(() => {
    const laid = layoutGraph(mergedGraph);
    setNodes((prev) => {
      const posById = new Map(prev.map((n) => [n.id, n.position]));
      return laid.nodes.map((n) => ({ ...n, position: posById.get(n.id) ?? n.position })) as CanvasFlowNode[];
    });
    setEdges(mergedGraph.edges as Edge[]);
    // eslint-disable-next-line react-hooks/exhaustive-deps -- intentionally keyed on topology only
  }, [topologySignature, setNodes, setEdges]);

  // Update node data (summary + dirty mark) in place, without moving nodes.
  useEffect(() => {
    const dataById = new Map(mergedGraph.nodes.map((n) => [n.id, n.data]));
    setNodes(
      (prev) => prev.map((n) => (dataById.has(n.id) ? { ...n, data: dataById.get(n.id)! } : n)) as CanvasFlowNode[],
    );
  }, [mergedGraph, setNodes]);

  const toggleConnections = useCallback(() => setShowConnections((v) => !v), []);

  // Re-run auto-layout: reset every node to its fresh dagre position and re-fit.
  const autoLayout = useCallback(() => {
    const laid = layoutGraph(mergedGraph);
    setNodes(laid.nodes as CanvasFlowNode[]);
    requestAnimationFrame(() => fitView(FIT_OPTIONS));
  }, [mergedGraph, setNodes, fitView]);

  const onNodeClick = useCallback<NodeMouseHandler<CanvasFlowNode>>(
    (_event, node) => {
      if (node.type !== "resource") return; // attachment node — display-only, no drawer
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

  const [addVolumeOpen, setAddVolumeOpen] = useState(false);
  const [addVolumeResourceIdx, setAddVolumeResourceIdx] = useState<number | null>(null);

  /** Apply a pure draft mutation, starting a session lazily when needed. */
  const applyDraft = useCallback(
    (fn: (draft: EditSessionDraft) => EditSessionDraft) => {
      const current: EditSessionDraft = session.isActive
        ? { resources: session.draft.resources, volumes: session.draft.volumes }
        : { resources: draftResources, volumes: draftVolumes };
      const next = fn(current);
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          { linkedAddonIds: new Set(connectionAddonIds), draft: next },
        );
      } else {
        session.updateResources(() => next.resources);
        session.updateVolumes(() => next.volumes);
      }
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  const onCreateVolume = useCallback(
    (input: { name: string; size: string; resourceIdx: number; targetPath: string }) => {
      applyDraft((draft) => ({
        resources: addMount(draft.resources, input.resourceIdx, {
          volumeName: input.name,
          targetPath: input.targetPath,
        }),
        volumes: [...draft.volumes, newVolume({ name: input.name, size: input.size })],
      }));
    },
    [applyDraft],
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
          canAddVolume={resources.length > 0}
          onAddVolume={() => setAddVolumeOpen(true)}
        />
      </div>
      <DrawerStack
        panels={panels}
        front={frontBody}
        onTruncate={truncateDrawers}
        onPop={popDrawer}
        onCloseAll={closeAllDrawers}
      />
      <AddVolumeDialog
        open={addVolumeOpen}
        onOpenChange={(o) => {
          setAddVolumeOpen(o);
          if (!o) setAddVolumeResourceIdx(null);
        }}
        resources={resources}
        volumes={volumes}
        initialResourceIdx={addVolumeResourceIdx}
        onCreate={onCreateVolume}
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
