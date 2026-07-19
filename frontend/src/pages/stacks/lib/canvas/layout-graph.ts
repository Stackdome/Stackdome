import dagre from "dagre";
import type { CanvasGraph } from "./graph-from-connections";

/** Card dimensions dagre reserves per node (matches ResourceNode's box). */
export const NODE_WIDTH = 216;
export const NODE_HEIGHT = 88;

/** Horizontal pitch between consecutive freshly-added nodes so they don't stack. */
export const NEW_NODE_GAP_X = NODE_WIDTH + NODE_HEIGHT;
/** Vertical drop below the preserved layout where new nodes are parked. */
export const NEW_NODE_OFFSET_Y = NODE_HEIGHT * 2;

export interface LayoutOptions {
  direction?: "LR" | "TB" | "BT";
}

/** Vertical gap between ranks (tiers) — room for bezier fans to separate. */
export const RANK_SEP = 140;
/** Horizontal gap between siblings within a rank. */
export const NODE_SEP = 64;

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
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  }
  for (const edge of graph.edges) {
    g.setEdge(edge.source, edge.target);
  }

  dagre.layout(g);

  const nodes = graph.nodes.map((node) => {
    const { x, y } = g.node(node.id);
    return { ...node, position: { x: x - NODE_WIDTH / 2, y: y - NODE_HEIGHT / 2 } };
  });

  return { nodes, edges: graph.edges };
}
