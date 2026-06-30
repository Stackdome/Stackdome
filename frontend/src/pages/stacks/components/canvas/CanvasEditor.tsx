import { ReactFlow, Background, Controls, type Node, type Edge } from "@xyflow/react";
import "@xyflow/react/dist/style.css";

interface CanvasEditorProps {
  nodes: Node[];
  edges: Edge[];
}

/**
 * Thin React Flow shell. It owns only the render surface — nodes and edges are
 * derived upstream from the edit session (pure calculations), so this component
 * stays a dumb view. Slice 1 adds custom node types and selection.
 */
export function CanvasEditor({ nodes, edges }: CanvasEditorProps) {
  return (
    <div className="h-full w-full" data-testid="stack-canvas">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        fitView
        colorMode="system"
        proOptions={{ hideAttribution: true }}
      >
        <Background />
        <Controls />
      </ReactFlow>
    </div>
  );
}
