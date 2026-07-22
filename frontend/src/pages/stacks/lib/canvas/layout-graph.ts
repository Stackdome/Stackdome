import dagre from "dagre";
import type { CanvasGraph, CanvasNode, ResourceNodeData } from "./graph-from-connections";

/** Card dimensions dagre reserves per node (matches ResourceNode's box —
 *  header + summary + optional port detail line). */
export const NODE_WIDTH = 216;
export const NODE_HEIGHT = 104;

/** Attachment (volume/secret/object-store) card box — w-[180px], two compact
 *  rows. Used as the collision-box fallback for unmeasured attachment nodes. */
export const ATTACHMENT_NODE_WIDTH = 180;
export const ATTACHMENT_NODE_HEIGHT = 56;

export interface LayoutOptions {
  direction?: "LR" | "TB" | "BT";
}

/** Vertical gap between ranks (tiers) — room for bezier fans to separate. */
export const RANK_SEP = 140;
/** Horizontal gap between siblings within a rank. */
export const NODE_SEP = 64;

/** Card growth per extra port detail line beyond the first. */
const PORT_LINE_HEIGHT = 18;
/** Card growth per docked volume chip row. */
const VOLUME_ROW_HEIGHT = 30;

/** Box dagre reserves per node — matches what the card actually renders
 *  (extra port lines and docked volume rows grow resource cards; attachment
 *  nodes are the smaller fixed card). */
function nodeLayoutSize(node: CanvasNode): { width: number; height: number } {
  if (node.type === "attachment") {
    return { width: ATTACHMENT_NODE_WIDTH, height: ATTACHMENT_NODE_HEIGHT };
  }
  const data = node.data as ResourceNodeData;
  const extraPortLines = Math.max(0, (data.details?.length ?? 0) - 1);
  const volumeRows = data.volumes?.length ?? 0;
  return {
    width: NODE_WIDTH,
    height: NODE_HEIGHT + extraPortLines * PORT_LINE_HEIGHT + volumeRows * VOLUME_ROW_HEIGHT,
  };
}

/**
 * Pure auto-layout. Runs dagre over the graph and returns a NEW graph with
 * resolved positions (copy-on-write — the input is never mutated). Deterministic
 * for a given input, so unrelated re-renders never reshuffle the board.
 *
 * dagre reports node centres; React Flow positions are top-left, so we offset by
 * half the node box.
 */
export function layoutGraph(graph: CanvasGraph, options: LayoutOptions = {}): CanvasGraph {
  if (graph.nodes.length === 0) return { nodes: [], edges: graph.edges };

  const g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: options.direction ?? "BT", nodesep: NODE_SEP, ranksep: RANK_SEP, marginx: 24, marginy: 24 });
  g.setDefaultEdgeLabel(() => ({}));

  for (const node of graph.nodes) {
    g.setNode(node.id, nodeLayoutSize(node));
  }
  for (const edge of graph.edges) {
    g.setEdge(edge.source, edge.target);
  }

  dagre.layout(g);

  const nodes = graph.nodes.map((node) => {
    const { x, y } = g.node(node.id);
    const { width, height } = nodeLayoutSize(node);
    return { ...node, position: { x: x - width / 2, y: y - height / 2 } };
  });

  return { nodes, edges: graph.edges };
}
