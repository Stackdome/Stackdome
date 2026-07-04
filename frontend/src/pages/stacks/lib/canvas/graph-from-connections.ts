import type { FormStackResourceData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import { nodePresentation, type GlyphKind, type DotState } from "./node-presentation";
import { splitEnvRows, type FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import type { StackConnection } from "@/api/connections";

/**
 * Canvas node kinds. Workload kinds (service, addon) plus the compact
 * "attachment" kinds (secret, volume, object store) that connections can
 * reference as producers/consumers. Worker/cron nodes are intentionally
 * absent until the WorkloadType CRD is wired into the api-server.
 */
export const NODE_KIND = {
  service: "service",
  addon: "addon",
  secret: "secret",
  volume: "volume",
  objectStore: "objectStore",
} as const;
export type NodeKind = (typeof NODE_KIND)[keyof typeof NODE_KIND];
export type AttachmentKind = typeof NODE_KIND.secret | typeof NODE_KIND.volume | typeof NODE_KIND.objectStore;

/** Canvas edge kinds — mirrors StackConnection["kind"] plus the derived depends_on edge. */
export const EDGE_KIND = {
  env: "env",
  volumeMount: "volume_mount",
  buildArtifactSource: "build_artifact_source",
  dependsOn: "depends_on",
} as const;
export type EdgeKind = (typeof EDGE_KIND)[keyof typeof EDGE_KIND];

/** Whether an edge reflects an authored StackConnection or is derived (e.g. depends_on ordering). */
export const EDGE_SOURCE_OF_TRUTH = { connection: "connection", derived: "derived" } as const;
export type EdgeSourceOfTruth = (typeof EDGE_SOURCE_OF_TRUTH)[keyof typeof EDGE_SOURCE_OF_TRUTH];

/** Node id prefixes keep resource/addon/secret/volume/object-store namespaces from colliding. */
export const NODE_ID_PREFIX = {
  resource: "resource:",
  addon: "addon:",
  secret: "secret:",
  volume: "volume:",
  objectstore: "objectstore:",
} as const;

/** Attachment kind → the label shown on its compact node card. Exported for Task 3's drawer chrome. */
export const ATTACHMENT_LABEL: Record<AttachmentKind, string> = {
  [NODE_KIND.secret]: "SECRET",
  [NODE_KIND.volume]: "VOLUME",
  [NODE_KIND.objectStore]: "OBJECT STORE",
};

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
  /** Transient: true while a dragged volume hovers this card as a valid drop target. */
  dropTarget?: boolean;
  [key: string]: unknown; // satisfies React Flow's Record<string, unknown> node data
}

export interface AttachmentNodeData {
  kind: AttachmentKind;
  name: string;
  kindLabel: string; // "SECRET" | "VOLUME" | "OBJECT STORE"
  [key: string]: unknown;
}

export interface CanvasNode {
  id: string;
  type: "resource" | "attachment";
  data: ResourceNodeData | AttachmentNodeData;
  position: { x: number; y: number };
}

export interface ConnectionEdgeData {
  kind: EdgeKind;
  sourceOfTruth: EdgeSourceOfTruth;
  [key: string]: unknown;
}

export interface CanvasEdge {
  id: string;
  source: string;
  target: string;
  type: "connection";
  data: ConnectionEdgeData;
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
  /** Addon ids linked in the baseline — anything not here is a "new" link. */
  baselineAddonIds?: ReadonlySet<string>;
}

export interface DeriveGraphInput {
  resources: Partial<FormStackResourceData>[];
  linkedAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
  secretNameById: ReadonlyMap<string, string>;
  /** All volume names in the draft — only UNMOUNTED ones render as free attachment nodes. */
  volumeNames?: readonly string[];
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

function volumeChips(
  resource: Partial<FormStackResourceData>,
  knownVolumes: ReadonlySet<string> | null,
): VolumeChip[] {
  return (resource.volume_mounts ?? [])
    .filter((m) => !knownVolumes || knownVolumes.has((m.source_volume_name ?? "") as string))
    .map((m) => ({
      name: (m.source_volume_name ?? "") as string,
      mountPath: m.target_path as string | undefined,
    }));
}

export function edgeKey(kind: string, source: string, target: string): string {
  return `${kind}:${source}->${target}`;
}

function nodeIdOfConnRef(ref: StackConnection["from"] | undefined): string | null {
  if (!ref) return null;
  switch (ref.type) {
    case "stack_resource":
      return ref.name ? NODE_ID_PREFIX.resource + ref.name : null;
    case "addon/postgres":
      return ref.id ? NODE_ID_PREFIX.addon + ref.id : null;
    case "secret":
      return ref.id ? NODE_ID_PREFIX.secret + ref.id : null;
    case "volume":
      return ref.name ? NODE_ID_PREFIX.volume + ref.name : null;
    case "object_store":
      return (ref.name ?? ref.id) ? NODE_ID_PREFIX.objectstore + (ref.name ?? ref.id) : null;
    default:
      return null;
  }
}

/**
 * Pure projection of the edit-session draft into a node/edge graph, sourced
 * from real StackConnection semantics rather than inferring edges from raw
 * env rows.
 *
 * - service node per resource, addon node per linked addon id
 * - secret attachment nodes created on demand from connection producers
 * - edges projected from each resource's authored env connections
 *   (`splitEnvRows`), plus derived `depends_on` edges
 * - MOUNTED volumes render only as chips docked on their owning resource
 *   card (no separate node, no edge); UNMOUNTED volumes render as free
 *   attachment nodes so they stay visible and re-attachable
 *
 * Positions are left at the origin; `layoutGraph` assigns real coordinates.
 * No React/React Flow imports — this is unit-testable in isolation.
 */
export function deriveGraph(input: DeriveGraphInput): CanvasGraph {
  const nodes: CanvasNode[] = [];

  // null = caller didn't supply volumeNames (tests, older callers): no filtering.
  const knownVolumes = input.volumeNames ? new Set(input.volumeNames) : null;

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
        volumes: volumeChips(resource, knownVolumes),
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

  const nodeById = new Map(nodes.map((n) => [n.id, n]));
  const edges: CanvasEdge[] = [];
  const seen = new Set<string>();

  const ensureAttachment = (id: string, kind: AttachmentKind, name: string) => {
    if (nodeById.has(id)) return;
    const node: CanvasNode = {
      id,
      type: "attachment",
      position: { x: 0, y: 0 },
      data: { kind, name, kindLabel: ATTACHMENT_LABEL[kind] },
    };
    nodeById.set(id, node);
    nodes.push(node);
  };

  const addEdge = (
    kind: EdgeKind,
    sourceOfTruth: EdgeSourceOfTruth,
    source: string | null,
    target: string | null,
  ) => {
    if (!source || !target || source === target) return;
    if (!nodeById.has(source) || !nodeById.has(target)) return;
    const id = edgeKey(kind, source, target);
    if (seen.has(id)) return;
    seen.add(id);
    edges.push({ id, source, target, type: "connection", data: { kind, sourceOfTruth } });
  };

  for (const resource of input.resources) {
    const name = resource.name ?? "";
    const conns: StackConnection[] = splitEnvRows(name, envVarsOf(resource) as FormEnvRow[]).connections;
    for (const conn of conns) {
      // Secret producers materialize as attachment nodes on demand.
      if (conn.from?.type === "secret" && conn.from.id) {
        ensureAttachment(
          NODE_ID_PREFIX.secret + conn.from.id,
          NODE_KIND.secret,
          input.secretNameById.get(conn.from.id) ?? conn.from.id,
        );
      }
      addEdge(conn.kind as EdgeKind, EDGE_SOURCE_OF_TRUTH.connection, nodeIdOfConnRef(conn.from), nodeIdOfConnRef(conn.to));
    }
  }

  // Unmounted volumes float as free attachment nodes; mounted ones are
  // already represented by the chip on their owning resource card.
  const mountedVolumes = new Set<string>();
  for (const resource of input.resources) {
    for (const m of resource.volume_mounts ?? []) {
      if (m.source_volume_name) mountedVolumes.add(m.source_volume_name as string);
    }
  }
  for (const volumeName of input.volumeNames ?? []) {
    if (!volumeName || mountedVolumes.has(volumeName)) continue;
    ensureAttachment(NODE_ID_PREFIX.volume + volumeName, NODE_KIND.volume, volumeName);
  }

  for (const resource of input.resources) {
    for (const dep of resource.depends_on ?? []) {
      addEdge(
        EDGE_KIND.dependsOn,
        EDGE_SOURCE_OF_TRUTH.derived,
        NODE_ID_PREFIX.resource + dep,
        NODE_ID_PREFIX.resource + (resource.name ?? ""),
      );
    }
  }

  return { nodes, edges };
}
