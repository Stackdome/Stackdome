import type { FormStackResourceData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import { nodePresentation, type GlyphKind, type DotState } from "./node-presentation";

/**
 * Canvas node kinds. Only the workload kinds that exist in the backend today —
 * services (stack resources) and managed postgres addons. Worker/cron nodes are
 * intentionally absent until the WorkloadType CRD is wired into the api-server.
 */
export const NODE_KIND = { service: "service", addon: "addon" } as const;
export type NodeKind = (typeof NODE_KIND)[keyof typeof NODE_KIND];

/** Node id prefixes keep resource and addon namespaces from colliding. */
const NODE_ID_PREFIX = { resource: "resource:", addon: "addon:" } as const;

export interface VolumeChip {
  name: string;
  mountPath?: string;
}

/** Unsaved-change state shown as a mark on the node card. */
export type DirtyState = "new" | "edited" | "removed";

export interface ResourceNodeData {
  kind: NodeKind;
  name: string;
  /** Role/tech label shown as the node's top-right badge (Web/Redis/Postgres…). */
  kindLabel: string;
  /** Glyph identifier for the node icon. */
  glyph: GlyphKind;
  /** Status-dot colour bucket. */
  dotState: DotState;
  summary: string;
  status?: string;
  volumes: VolumeChip[];
  dirtyState?: DirtyState;
  /** Index into the edit session's resource array; absent for addon nodes. */
  resourceIdx?: number;
  [key: string]: unknown; // satisfies React Flow's Record<string, unknown> node data
}

export interface CanvasNode {
  id: string;
  type: "resource";
  data: ResourceNodeData;
  position: { x: number; y: number };
}

export interface CanvasEdge {
  id: string;
  source: string;
  target: string;
}

export interface CanvasGraph {
  nodes: CanvasNode[];
  edges: CanvasEdge[];
}

/** Optional unsaved-change context used to mark nodes new/edited/removed. */
export interface DirtyInput {
  /** Indices of resources that differ from baseline. */
  dirtyResourceIdx?: ReadonlySet<number>;
  /** Resource count in the baseline — anything at or beyond it is "new". */
  baselineResourceCount?: number;
  /** Addon ids queued for removal. */
  pendingDetach?: ReadonlySet<string>;
  /** Addon ids linked in the baseline — anything not here is a "new" link. */
  baselineAddonIds?: ReadonlySet<string>;
}

export interface DeriveGraphInput {
  resources: Partial<FormStackResourceData>[];
  linkedAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
  dirty?: DirtyInput;
}

function serviceDirtyState(idx: number, dirty: DirtyInput | undefined): DirtyState | undefined {
  if (!dirty) return undefined;
  if (dirty.baselineResourceCount != null && idx >= dirty.baselineResourceCount) return "new";
  if (dirty.dirtyResourceIdx?.has(idx)) return "edited";
  return undefined;
}

function addonDirtyState(addonId: string, dirty: DirtyInput | undefined): DirtyState | undefined {
  if (!dirty) return undefined;
  if (dirty.pendingDetach?.has(addonId)) return "removed";
  if (dirty.baselineAddonIds != null && !dirty.baselineAddonIds.has(addonId)) return "new";
  return undefined;
}

function resourceNodeId(name: string): string {
  return NODE_ID_PREFIX.resource + name;
}

function addonNodeId(addonId: string): string {
  return NODE_ID_PREFIX.addon + addonId;
}

function envVarsOf(resource: Partial<FormStackResourceData>): FormEnvVarData[] {
  return (resource.execution_config?.environment_variables ?? []) as FormEnvVarData[];
}

function servicePresentation(resource: Partial<FormStackResourceData>) {
  return nodePresentation({
    isAddon: false,
    image: resource.image_spec?.image,
    hasBuild: !!resource.build_spec,
    ports: (resource.ports ?? []).map((p) => ({
      number: p.number,
      protocol: p.protocol,
      exposedToPublic: p.exposed_to_public,
    })),
  });
}

function volumeChips(resource: Partial<FormStackResourceData>): VolumeChip[] {
  return (resource.volume_mounts ?? []).map((m) => ({
    name: (m.name ?? "") as string,
    mountPath: m.mount_path as string | undefined,
  }));
}

/**
 * Pure projection of the edit-session draft into a node/edge graph.
 *
 * - service node per resource, addon node per linked addon id
 * - edges from connection env-var references (`from: "addon" | "resource"`),
 *   kept only when the target node actually exists, and deduped
 * - volumes folded into their owning node as chips
 *
 * Positions are left at the origin; `layoutGraph` assigns real coordinates.
 * No React/React Flow imports — this is unit-testable in isolation.
 */
export function deriveGraph(input: DeriveGraphInput): CanvasGraph {
  const nodes: CanvasNode[] = [];

  input.resources.forEach((resource, idx) => {
    const name = resource.name ?? "";
    const pres = servicePresentation(resource);
    nodes.push({
      id: resourceNodeId(name),
      type: "resource",
      position: { x: 0, y: 0 },
      data: {
        kind: NODE_KIND.service,
        name,
        kindLabel: pres.kindLabel,
        glyph: pres.glyph,
        dotState: pres.dotState,
        summary: pres.summary,
        status: resource.status as string | undefined,
        volumes: volumeChips(resource),
        dirtyState: serviceDirtyState(idx, input.dirty),
        resourceIdx: idx,
      },
    });
  });

  for (const addonId of input.linkedAddonIds) {
    const pres = nodePresentation({ isAddon: true });
    nodes.push({
      id: addonNodeId(addonId),
      type: "resource",
      position: { x: 0, y: 0 },
      data: {
        kind: NODE_KIND.addon,
        name: input.addonNameById.get(addonId) ?? addonId,
        kindLabel: pres.kindLabel,
        glyph: pres.glyph,
        dotState: pres.dotState,
        summary: pres.summary,
        volumes: [],
        dirtyState: addonDirtyState(addonId, input.dirty),
      },
    });
  }

  const nodeIds = new Set(nodes.map((n) => n.id));
  const edges: CanvasEdge[] = [];
  const seenEdgeIds = new Set<string>();

  const addEdge = (source: string, target: string) => {
    if (!nodeIds.has(target) || source === target) return;
    const id = `${source}->${target}`;
    if (seenEdgeIds.has(id)) return;
    seenEdgeIds.add(id);
    edges.push({ id, source, target });
  };

  for (const resource of input.resources) {
    const source = resourceNodeId(resource.name ?? "");
    for (const envVar of envVarsOf(resource)) {
      if (envVar.from === "addon") addEdge(source, addonNodeId(envVar.addonId));
      else if (envVar.from === "resource") addEdge(source, resourceNodeId(envVar.resourceName));
    }
  }

  return { nodes, edges };
}
