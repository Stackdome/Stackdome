import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import {
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  useNodesInitialized,
  useReactFlow,
  type Edge,
  type NodeMouseHandler,
  type OnNodeDrag,
  type XYPosition,
} from "@xyflow/react";
import type {
  FormStackResourceData,
  FormVolumeExtendedData as VolumeFormData,
} from "@/pages/stacks/schemas/form-schema";
import type { EditSessionDraft, EditSessionTab, UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import { useStackTopology } from "@/pages/stacks/hooks/use-stack-topology";
import type { ReleaseLiveStatus } from "@/api/releases";
import type { PublicEndpoint } from "@/pages/stacks/components/editor/public-endpoint-row";
import { deriveGraph } from "@/pages/stacks/lib/canvas/graph-from-connections";
import { mergeTopology } from "@/pages/stacks/lib/canvas/merge-topology";
import {
  layoutGraph,
  ATTACHMENT_NODE_HEIGHT,
  ATTACHMENT_NODE_WIDTH,
  NODE_HEIGHT,
  NODE_SEP,
  NODE_WIDTH,
} from "@/pages/stacks/lib/canvas/layout-graph";
import { resolveCollisions, type CollidableNode } from "@/pages/stacks/lib/canvas/resolve-collisions";
import { carryPositions } from "@/pages/stacks/lib/canvas/carry-positions";

/** Collision-box dims for nodes React Flow hasn't measured yet (fresh layout
 *  output) — attachment cards are markedly smaller than resource cards. */
function attachmentAwareFallbackSize(node: CollidableNode): { width: number; height: number } {
  return (node as { type?: string }).type === "attachment"
    ? { width: ATTACHMENT_NODE_WIDTH, height: ATTACHMENT_NODE_HEIGHT }
    : { width: NODE_WIDTH, height: NODE_HEIGHT };
}
import { addBlockToStack } from "@/pages/stacks/lib/block-to-form";
import { blockCatalog, getBlockById } from "@/pages/stacks/data/blocks/registry";
import { CanvasEditor } from "./canvas-editor";
import { ResourceDrawer } from "./resource-drawer";
import { VolumeDrawer } from "./volume-drawer";
import { AddVolumeDialog } from "./add-volume-dialog";
import { CanvasContextMenu, type CanvasMenuTarget } from "./canvas-context-menu";
import { MountPathDialog } from "./mount-path-dialog";
import { addMount, newVolume, removeMountsOf } from "@/pages/stacks/lib/canvas/volume-ops";
import {
  NODE_KIND,
  NODE_ID_PREFIX,
  type AttachmentNodeData,
  type ResourceNodeData,
} from "@/pages/stacks/lib/canvas/graph-from-connections";
import { useConfirm } from "@/components/branded/confirm";
import { DrawerStack, type DrawerPanelDescriptor } from "./drawer-stack";
import {
  replaceStack,
  pushEntry,
  truncateTo,
  popEntry,
  entryKey,
  type DrawerEntry,
} from "@/pages/stacks/lib/canvas/drawer-stack";
import { computeDrawerInset } from "@/pages/stacks/lib/canvas/drawer-inset";
import { nodePresentation } from "@/pages/stacks/lib/canvas/node-presentation";
import { NodeGlyph } from "./nodes/node-glyph";
import { HardDrive } from "lucide-react";
import { FIT_OPTIONS } from "./fit-options";
import type { CanvasFlowNode } from "./canvas-editor";

interface ArchitectureTabProps {
  session: UseStackEditSession;
  /** Diff baseline (the deployed release snapshot when one exists). */
  baselineResources: Partial<FormStackResourceData>[];
  baselineVolumes: Partial<VolumeFormData>[];
  /** Current server state — what the canvas shows when no session is active,
   *  and what a lazily-started session's working draft seeds from. */
  draftResources: Partial<FormStackResourceData>[];
  draftVolumes: Partial<VolumeFormData>[];
  /** Server-computed resource outputs keyed by resource name — used to populate
   *  the env-var output pickers for resources whose draft copy has no outputs. */
  serverOutputsByName?: ReadonlyMap<string, string[]>;
  connectionAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
  /** addonId → live state (e.g. "Ready"), for canvas addon node dots. */
  addonStateById?: ReadonlyMap<string, string>;
  errors: { [index: number]: { [field: string]: string | undefined } };
  /** Switch the editor to the Logs tab (from the drawer's "View logs"),
   *  pre-filtering to the named resource when given. */
  onViewLogs?: (resourceName?: string) => void;
  /** Null for draft (unsaved) stacks — no server topology exists yet. */
  topologyIds: { orgId: string; projectName: string; stackId: string } | null;
  /** Bump to force a topology refetch (wired to autosave refreshes). */
  topologyRefreshKey: number;
  /** Immediate, confirm-gated server-side volume deletion. Undefined for draft
   *  (unsaved) stacks — nothing exists server-side to delete yet. */
  onDeleteVolume?: (name: string) => Promise<boolean>;
  /** Names of volumes that already exist server-side; their spec is immutable. */
  persistedVolumeNames?: ReadonlySet<string>;
  /** A release is in flight — node status dots show pending until it terminates
   *  (per-resource server state lags the deploy and would flash a stale Ready). */
  releaseInFlight?: boolean;
  /** Live per-resource status, keyed by resource name — from the status
   *  release's live_status.resources. Drives node dots + the resource drawer. */
  liveStatusResources?: ReleaseLiveStatus["resources"];
  /** Service → best live URL (as shown in the header's PUBLIC row); feeds the
   *  resource drawer's endpoint line. */
  publicEndpoints?: PublicEndpoint[];
  /** Bumped by the parent to request opening a resource drawer (banner "jump to
   *  error") on the tab holding the field; the nonce distinguishes repeat jumps
   *  to the same resource. */
  openResourceSignal?: { index: number; tab: EditSessionTab; nonce: number } | null;
}

function StackCanvasFlow({
  session,
  baselineResources,
  baselineVolumes,
  draftResources,
  draftVolumes,
  serverOutputsByName,
  connectionAddonIds,
  addonNameById,
  addonStateById,
  errors,
  onViewLogs,
  topologyIds,
  topologyRefreshKey,
  onDeleteVolume,
  persistedVolumeNames,
  releaseInFlight,
  liveStatusResources,
  publicEndpoints,
  openResourceSignal,
}: ArchitectureTabProps) {
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

  const { topology } = useStackTopology({ ids: topologyIds, refreshKey: topologyRefreshKey });

  // Local connection-derived data (cheap, pure). Re-runs on any edit.
  const dataGraph = useMemo(
    () => deriveGraph({ resources, linkedAddonIds, addonNameById, addonStateById, volumeNames, dirty }),
    [resources, linkedAddonIds, addonNameById, addonStateById, volumeNames, dirty],
  );
  // Live status keyed by canvas node id, so it wins over the topology
  // endpoint's own (possibly unscoped) per-node state.
  const liveStateByNodeId = useMemo(() => {
    const entries = Object.entries(liveStatusResources ?? {});
    if (entries.length === 0) return undefined;
    return new Map(entries.map(([name, s]) => [NODE_ID_PREFIX.resource + name, s.state]));
  }, [liveStatusResources]);

  // Local graph enhanced with server-derived edges + runtime status.
  const mergedGraphBare = useMemo(
    () => mergeTopology(dataGraph, topology, releaseInFlight, liveStateByNodeId),
    [dataGraph, topology, releaseInFlight, liveStateByNodeId],
  );
  // Overlay live per-port public URLs so a card's public port lines link out.
  const mergedGraph = useMemo(() => {
    if (!publicEndpoints?.length) return mergedGraphBare;
    const byService = new Map(publicEndpoints.map((e) => [e.service, e.urls]));
    return {
      ...mergedGraphBare,
      nodes: mergedGraphBare.nodes.map((n) => {
        const urls = n.type === "resource" ? byService.get((n.data as { name: string }).name) : undefined;
        if (!urls?.length) return n;
        const portUrls = Object.fromEntries(
          urls.filter((u) => u.target_port != null).map((u) => [u.target_port!, u.url]),
        );
        return { ...n, data: { ...n.data, portUrls } };
      }),
    };
  }, [mergedGraphBare, publicEndpoints]);
  // Signature of the node/edge id-set — changes only when topology changes.
  const topologySignature = useMemo(
    () => `${mergedGraph.nodes.map((n) => n.id).join("|")}::${mergedGraph.edges.map((e) => e.id).join("|")}`,
    [mergedGraph],
  );

  const [nodes, setNodes, onNodesChange] = useNodesState<CanvasFlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [showConnections, setShowConnections] = useState(true);
  const [drawerStack, setDrawerStack] = useState<DrawerEntry[]>([]);
  const [menuTarget, setMenuTarget] = useState<CanvasMenuTarget | null>(null);
  const confirm = useConfirm();
  const { fitView, getIntersectingNodes, getViewport, setViewport } = useReactFlow();
  const dragStartPos = useRef<XYPosition | null>(null);

  // One-time fit when the canvas first gains measured nodes: async imports and
  // stack loads populate nodes after mount, so ReactFlow's own fitView prop
  // fires against an empty canvas. Never re-fits — the hook re-flips when
  // nodes are added mid-edit, and a re-fit then would yank the viewport.
  const nodesInitialized = useNodesInitialized();
  const didInitialFit = useRef(false);
  useEffect(() => {
    if (!nodesInitialized || didInitialFit.current) return;
    didInitialFit.current = true;
    // One frame late: measurement flips `nodesInitialized` before the final
    // card sizes (attachment rows) and container height settle.
    requestAnimationFrame(() => fitView(FIT_OPTIONS));
  }, [nodesInitialized, fitView]);

  // When the drawer claims/releases horizontal space, the canvas container is
  // squeezed from the right. Pan the viewport by half that delta so the point
  // that was at the visible center stays centered — nodes glide left with the
  // drawer instead of sitting still while the container shrinks around them.
  const prevDrawerInsetRef = useRef(0);
  useEffect(() => {
    const inset = computeDrawerInset(drawerStack.length, window.innerWidth);
    const delta = inset - prevDrawerInsetRef.current;
    prevDrawerInsetRef.current = inset;
    if (delta === 0) return;
    const vp = getViewport();
    void setViewport({ ...vp, x: vp.x - delta / 2 }, { duration: 260 });
  }, [drawerStack.length, getViewport, setViewport]);

  const isFloatingVolume = (node: CanvasFlowNode) =>
    node.type === "attachment" && (node.data as AttachmentNodeData).kind === NODE_KIND.volume;

  /** First intersecting service card (addons have no resourceIdx and never qualify). */
  const dropTargetFor = useCallback(
    (node: CanvasFlowNode): CanvasFlowNode | null => {
      const hit = getIntersectingNodes(node).find(
        (n) => n.type === "resource" && (n.data as ResourceNodeData).resourceIdx != null,
      );
      return (hit as CanvasFlowNode) ?? null;
    },
    [getIntersectingNodes],
  );

  const onNodeDragStart = useCallback<OnNodeDrag<CanvasFlowNode>>((_event, node) => {
    dragStartPos.current = isFloatingVolume(node) ? { ...node.position } : null;
  }, []);

  const onNodeDrag = useCallback<OnNodeDrag<CanvasFlowNode>>(
    (_event, node) => {
      if (!isFloatingVolume(node)) return;
      const target = dropTargetFor(node);
      setNodes((prev) =>
        prev.map((n) => {
          const isTarget = n.id === target?.id;
          const current = (n.data as ResourceNodeData).dropTarget ?? false;
          if (current === isTarget) return n;
          return { ...n, data: { ...n.data, dropTarget: isTarget } } as CanvasFlowNode;
        }),
      );
    },
    [dropTargetFor, setNodes],
  );

  const onNodeDragStop = useCallback<OnNodeDrag<CanvasFlowNode>>(
    (_event, node) => {
      if (isFloatingVolume(node)) {
        const target = dropTargetFor(node);
        // Clear all rings.
        setNodes((prev) =>
          prev.map((n) =>
            (n.data as ResourceNodeData).dropTarget
              ? ({ ...n, data: { ...n.data, dropTarget: false } } as CanvasFlowNode)
              : n,
          ),
        );
        if (target) {
          const volumeName = node.id.slice(NODE_ID_PREFIX.volume.length);
          const resourceIdx = (target.data as ResourceNodeData).resourceIdx!;
          setAttachRequest({ volumeName, resourceIdx });
          return;
        }
      }
      // Plain reposition: only the dropped node yields to any overlap it
      // created — the rest of the user's arrangement stays pinned.
      dragStartPos.current = null;
      setNodes(
        (prev) =>
          resolveCollisions(prev, {
            margin: NODE_SEP / 2,
            isLocked: (n) => n.id !== node.id,
            fallbackSize: attachmentAwareFallbackSize,
          }) as CanvasFlowNode[],
      );
    },
    [dropTargetFor, setNodes],
  );

  const onNodeContextMenu = useCallback<NodeMouseHandler<CanvasFlowNode>>(
    (event, node) => {
      event.preventDefault();
      const { clientX: x, clientY: y } = event;
      const chipEl = (event.target as HTMLElement).closest("[data-volume-chip]");
      if (chipEl) {
        setMenuTarget({ kind: "volume-chip", volumeName: chipEl.getAttribute("data-volume-chip")!, x, y });
        return;
      }
      if (node.type === "attachment" && (node.data as AttachmentNodeData).kind === NODE_KIND.volume) {
        setMenuTarget({ kind: "volume-node", volumeName: (node.data as AttachmentNodeData).name, x, y });
        return;
      }
      if (node.type === "resource") {
        const data = node.data as ResourceNodeData;
        const idx = data.resourceIdx;
        if (idx != null) setMenuTarget({ kind: "resource", resourceIdx: idx, resourceName: data.name, x, y });
      }
    },
    [],
  );

  // Re-layout ONLY when topology changes; preserve in-session drag positions by id.
  useEffect(() => {
    const laid = layoutGraph(mergedGraph);
    setNodes((prev) => {
      // Empty canvas (no preserved nodes): use the fresh dagre layout as-is.
      if (prev.length === 0) return laid.nodes as CanvasFlowNode[];
      // carryPositions keeps renamed nodes in place (ids embed the name);
      // new nodes keep their dagre coords near their topological neighbours
      // and the locked collision pass shoves them clear of the frozen layout.
      const { nodes: next, keptIds } = carryPositions(prev, laid.nodes as CanvasFlowNode[]);
      return resolveCollisions(next, {
        margin: NODE_SEP / 2,
        isLocked: (n) => keptIds.has(n.id),
        fallbackSize: attachmentAwareFallbackSize,
      }) as CanvasFlowNode[];
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

  const openResourceDrawer = useCallback(
    (idx: number, tab: EditSessionTab = "configuration") => {
      // Activate an edit session lazily so drawer edits land in a draft.
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          {
            linkedAddonIds: new Set(connectionAddonIds),
            openResourceIdx: idx,
            openTab: tab,
            draft: { resources: draftResources, volumes: draftVolumes },
          },
        );
      } else {
        session.setOpenTab(tab);
      }
      setDrawerStack(replaceStack({ kind: "resource", index: idx }));
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  // Open the requested resource drawer when the parent bumps the signal (banner
  // jump). Gated on the nonce so an unrelated re-render can't reopen the drawer.
  const lastJumpNonceRef = useRef<number | null>(null);
  useEffect(() => {
    if (!openResourceSignal) return;
    if (lastJumpNonceRef.current === openResourceSignal.nonce) return;
    lastJumpNonceRef.current = openResourceSignal.nonce;
    openResourceDrawer(openResourceSignal.index, openResourceSignal.tab);
  }, [openResourceSignal, openResourceDrawer]);

  const openVolumeFromCanvas = useCallback(
    (volumeName: string) => {
      // The volume drawer reads session.draft — make sure a session exists first.
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          { linkedAddonIds: new Set(connectionAddonIds), draft: { resources: draftResources, volumes: draftVolumes } },
        );
      }
      setDrawerStack((s) => pushEntry(s, { kind: "volume", name: volumeName }));
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  const onNodeClick = useCallback<NodeMouseHandler<CanvasFlowNode>>(
    (event, node) => {
      if (node.type === "attachment") {
        const data = node.data as AttachmentNodeData;
        if (data.kind === NODE_KIND.volume) openVolumeFromCanvas(data.name);
        return; // secret/object-store attachments stay display-only
      }
      // A click on the attached-volume chip targets the volume, not the resource.
      const chipEl = (event.target as HTMLElement).closest("[data-volume-chip]");
      if (chipEl) {
        openVolumeFromCanvas(chipEl.getAttribute("data-volume-chip")!);
        return;
      }
      const idx = (node.data as ResourceNodeData).resourceIdx;
      if (idx == null) return; // addon node — managed via the Environment tab, no drawer in v1
      openResourceDrawer(idx);
    },
    [openResourceDrawer, openVolumeFromCanvas],
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

  /** Radix modal-from-menu race: opening a Dialog/AlertDialog in the same commit
   *  that closes the modal DropdownMenu makes the dialog save the menu's
   *  `pointer-events: none` body lock as the value to restore on close, wedging
   *  the page. Defer the open one tick so the menu's lock is released first. */
  const deferOpen = useCallback((fn: () => void) => {
    setTimeout(fn, 0);
  }, []);

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

  const onDisconnectVolume = useCallback(
    (volumeName: string) => {
      applyDraft((draft) => ({ ...draft, resources: removeMountsOf(draft.resources, volumeName) }));
    },
    [applyDraft],
  );

  const onDeleteVolumeConfirmed = useCallback(
    (volumeName: string) => {
      applyDraft((draft) => ({
        resources: removeMountsOf(draft.resources, volumeName),
        volumes: draft.volumes.filter((v) => v.name !== volumeName),
      }));
      // Saved stacks: destroy the volume server-side now (confirm dialog already
      // carried the data-loss warning). Draft stacks have nothing to delete yet —
      // onDeleteVolume is undefined and the local edit above is the whole story.
      void onDeleteVolume?.(volumeName);
    },
    [applyDraft, onDeleteVolume],
  );

  const onRequestDeleteVolume = useCallback(
    async (volumeName: string) => {
      const ok = await confirm(
        topologyIds == null
          ? {
            title: `Remove volume “${volumeName}”?`,
            confirmLabel: "Remove",
            variant: "destructive",
          }
          : {
            title: `Delete volume “${volumeName}”?`,
            description:
                "This immediately and permanently destroys the volume and all data stored on it. Any mounts are removed first. This cannot be undone.",
            confirmLabel: "Delete",
            variant: "destructive",
          },
      );
      if (!ok) return;
      onDeleteVolumeConfirmed(volumeName);
    },
    [confirm, topologyIds, onDeleteVolumeConfirmed],
  );

  const onRequestDeleteResource = useCallback(
    async (resourceName: string) => {
      const ok = await confirm({
        title: `Delete service “${resourceName}”?`,
        description:
          "The service and its configuration are removed when the stack deploys. This cannot be undone after deploy.",
        confirmLabel: "Delete",
        variant: "destructive",
      });
      if (!ok) return;
      applyDraft((draft) => ({ ...draft, resources: draft.resources.filter((r) => r.name !== resourceName) }));
      setDrawerStack([]);
    },
    [confirm, applyDraft],
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

  const [attachRequest, setAttachRequest] = useState<{ volumeName: string; resourceIdx: number | null } | null>(
    null,
  );

  const onAttachConfirm = useCallback(
    (input: { volumeName: string; resourceIdx: number; targetPath: string }) => {
      applyDraft((draft) => ({
        ...draft,
        resources: addMount(draft.resources, input.resourceIdx, {
          volumeName: input.volumeName,
          targetPath: input.targetPath,
        }),
      }));
      dragStartPos.current = null;
      setAttachRequest(null);
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
            image: r.source?.image?.ref,
            hasBuild: !!r.source?.git,
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
        serverOutputsByName={serverOutputsByName}
        connectionAddonIds={connectionAddonIds}
        errors={errors[frontEntry.index] ?? {}}
        onClose={popDrawer}
        onRemove={removeResource}
        onViewLogs={onViewLogs}
        onOpenVolume={openVolume}
        liveStatusResources={liveStatusResources}
        publicUrls={publicEndpoints?.find((e) => e.service === resources[frontEntry.index]?.name)?.urls}
      />
    ) : frontEntry ? (
      <VolumeDrawer
        key={entryKey(frontEntry)}
        volumeName={frontEntry.name}
        session={session}
        onClose={popDrawer}
        onRequestRemove={onRequestDeleteVolume}
        persisted={persistedVolumeNames?.has(frontEntry.name) ?? false}
      />
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
          onNodeContextMenu={onNodeContextMenu}
          onNodeDragStart={onNodeDragStart}
          onNodeDrag={onNodeDrag}
          onNodeDragStop={onNodeDragStop}
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
      <CanvasContextMenu
        target={menuTarget}
        onClose={() => setMenuTarget(null)}
        onOpenResource={openResourceDrawer}
        onAddVolumeToResource={(idx) => {
          deferOpen(() => {
            setAddVolumeResourceIdx(idx);
            setAddVolumeOpen(true);
          });
        }}
        onDeleteResource={onRequestDeleteResource}
        onDisconnectVolume={onDisconnectVolume}
        onOpenVolume={openVolumeFromCanvas}
        onRequestDeleteVolume={onRequestDeleteVolume}
        onRequestAttach={(volumeName) => deferOpen(() => setAttachRequest({ volumeName, resourceIdx: null }))}
      />
      <MountPathDialog
        volumeName={attachRequest?.volumeName ?? null}
        resources={resources}
        resourceIdx={attachRequest?.resourceIdx ?? null}
        onCancel={() => {
          const req = attachRequest;
          setAttachRequest(null);
          const start = dragStartPos.current;
          dragStartPos.current = null;
          if (req && start) {
            setNodes((prev) =>
              prev.map((n) => (n.id === NODE_ID_PREFIX.volume + req.volumeName ? { ...n, position: start } : n)),
            );
          }
        }}
        onAttach={onAttachConfirm}
      />
    </div>
  );
}

/**
 * Flag-gated entry mounted in the Configuration tab. Renders the stack as a node
 * graph; clicking a service node opens the config drawer. The `relative` wrapper
 * anchors the drawer overlay.
 */
export function ArchitectureTab(props: ArchitectureTabProps) {
  return (
    // Edge-to-edge inside the full-bleed editor shell (shell owns the chrome).
    <div className="relative h-full w-full overflow-hidden">
      <ReactFlowProvider>
        <StackCanvasFlow {...props} />
      </ReactFlowProvider>
    </div>
  );
}
