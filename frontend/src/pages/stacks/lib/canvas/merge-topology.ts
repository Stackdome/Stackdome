// Pure merge of the local (instant, authored) graph with server topology
// (derived edges + runtime state). Local is never mutated.
import type { StackTopology, TopologyNodeRef } from "@/api/topology";
import { statusVariant } from "@/components/branded/status-variant";
import {
  ATTACHMENT_LABEL, EDGE_SOURCE_OF_TRUTH, NODE_ID_PREFIX, NODE_KIND, edgeKey,
  type AttachmentKind, type CanvasGraph, type CanvasNode, type EdgeKind, type ResourceNodeData,
} from "./graph-from-connections";

// Secrets are intentionally absent: secret references never render on the
// canvas, so server-derived secret endpoints (and their edges) are skipped.
const ATTACHMENT_BY_REF_TYPE: Partial<Record<string, { kind: AttachmentKind; prefix: string }>> = {
  volume: { kind: NODE_KIND.volume, prefix: NODE_ID_PREFIX.volume },
  object_store: { kind: NODE_KIND.objectStore, prefix: NODE_ID_PREFIX.objectstore },
};

function nodeIdOfRef(ref: TopologyNodeRef | undefined): string | null {
  if (!ref) return null;
  switch (ref.type) {
    case "stack_resource": return ref.name ? NODE_ID_PREFIX.resource + ref.name : null;
    case "addon/postgres": return ref.id ? NODE_ID_PREFIX.addon + ref.id : null;
    case "secret":         return ref.id ? NODE_ID_PREFIX.secret + ref.id : null;
    case "volume":         return ref.name ? NODE_ID_PREFIX.volume + ref.name : null;
    case "object_store":   return (ref.name ?? ref.id) ? NODE_ID_PREFIX.objectstore + (ref.name ?? ref.id) : null;
    default: return null;
  }
}

export function mergeTopology(
  local: CanvasGraph,
  server: StackTopology | null | undefined,
  releaseInFlight = false,
): CanvasGraph {
  const merged = server ? mergeServer(local, server) : local;
  return releaseInFlight ? overlayReleaseInFlight(merged) : merged;
}

/**
 * While a release is in flight, per-resource server state lags the deploy (the
 * old workload keeps reporting Ready until the agent starts the rollout), so a
 * stale green would flash right after Deploy. Force resource dots to pending
 * until the release terminates; an already-reported error still wins.
 */
function overlayReleaseInFlight(graph: CanvasGraph): CanvasGraph {
  let changed = false;
  const nodes = graph.nodes.map((node) => {
    if (node.type !== "resource") return node;
    const current = (node.data as ResourceNodeData).dotVariant;
    if (current === "error" || current === "pending") return node;
    changed = true;
    return { ...node, data: { ...node.data, dotVariant: "pending" } } as CanvasNode;
  });
  return changed ? { nodes, edges: graph.edges } : graph;
}

function mergeServer(local: CanvasGraph, server: StackTopology): CanvasGraph {
  const nodes = [...local.nodes];
  const nodeById = new Map(nodes.map((n) => [n.id, n]));
  const serverNodeByListId = new Map(
    (server.nodes ?? []).map((n) => [nodeIdOfRef(n.ref), n] as const).filter(([id]) => id !== null),
  );
  let changed = false;

  // Runtime-state overlay: the fresher server topology state wins over the
  // dot colour derived at graph-build time.
  for (const [id, serverNode] of serverNodeByListId) {
    if (!serverNode.state) continue;
    const idx = nodes.findIndex((n) => n.id === id && n.type === "resource");
    if (idx === -1) continue;
    const node = nodes[idx];
    const dotVariant = statusVariant("resource", serverNode.state);
    if ((node.data as ResourceNodeData).dotVariant === dotVariant) continue;
    const next: CanvasNode = { ...node, data: { ...node.data, dotVariant } };
    nodes[idx] = next;
    nodeById.set(id!, next);
    changed = true;
  }

  // Union server-derived edges; local authored set is authoritative for the rest.
  const edges = [...local.edges];
  const seen = new Set(edges.map((e) => e.id));
  const ensureAttachment = (id: string, refType: string, label: string): boolean => {
    if (nodeById.has(id)) return true;
    const meta = ATTACHMENT_BY_REF_TYPE[refType];
    if (!meta) return false;
    const node: CanvasNode = {
      id, type: "attachment", position: { x: 0, y: 0 },
      data: { kind: meta.kind, name: label, kindLabel: ATTACHMENT_LABEL[meta.kind] },
    };
    nodes.push(node);
    nodeById.set(id, node);
    changed = true;
    return true;
  };

  for (const edge of server.edges ?? []) {
    if (edge.source_of_truth !== EDGE_SOURCE_OF_TRUTH.derived) continue;
    const source = nodeIdOfRef(edge.source);
    const target = nodeIdOfRef(edge.target);
    if (!source || !target || source === target) continue;
    const endpointOk = (id: string, ref: TopologyNodeRef) =>
      nodeById.has(id) ||
      ensureAttachment(id, ref.type, serverNodeByListId.get(id)?.label ?? ref.name ?? ref.id ?? id);
    if (!endpointOk(source, edge.source) || !endpointOk(target, edge.target)) continue;
    const id = edgeKey(edge.kind, source, target);
    if (seen.has(id)) continue;
    seen.add(id);
    edges.push({ id, source, target, type: "connection", data: { kind: edge.kind as EdgeKind, sourceOfTruth: EDGE_SOURCE_OF_TRUTH.derived } });
    changed = true;
  }

  return changed ? { nodes, edges } : local;
}
