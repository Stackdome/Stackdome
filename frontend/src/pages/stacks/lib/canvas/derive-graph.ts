import type { FormStackResourceData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

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

export interface ResourceNodeData {
  kind: NodeKind;
  name: string;
  summary: string;
  status?: string;
  volumes: VolumeChip[];
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

export interface DeriveGraphInput {
  resources: Partial<FormStackResourceData>[];
  linkedAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
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

function serviceSummary(resource: Partial<FormStackResourceData>): string {
  const image = resource.image_spec?.image;
  if (image) return image;
  if (resource.build_spec) return "git build";
  return "service";
}

function volumeChips(resource: Partial<FormStackResourceData>): VolumeChip[] {
  return (resource.volume_mounts ?? []).map((m) => ({
    name: m.name ?? "",
    mountPath: m.mount_path,
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
    nodes.push({
      id: resourceNodeId(name),
      type: "resource",
      position: { x: 0, y: 0 },
      data: {
        kind: NODE_KIND.service,
        name,
        summary: serviceSummary(resource),
        status: resource.status as string | undefined,
        volumes: volumeChips(resource),
        resourceIdx: idx,
      },
    });
  });

  for (const addonId of input.linkedAddonIds) {
    nodes.push({
      id: addonNodeId(addonId),
      type: "resource",
      position: { x: 0, y: 0 },
      data: {
        kind: NODE_KIND.addon,
        name: input.addonNameById.get(addonId) ?? addonId,
        summary: "postgres",
        volumes: [],
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
