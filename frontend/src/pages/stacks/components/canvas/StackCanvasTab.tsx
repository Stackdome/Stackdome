import { ReactFlowProvider } from "@xyflow/react";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import { CanvasEditor } from "./CanvasEditor";

interface StackCanvasTabProps {
  session: UseStackEditSession;
}

/**
 * Flag-gated entry point mounted inside the Configuration tab. Owns the React
 * Flow provider boundary. Slice 1 derives nodes/edges from `session`; for now
 * it renders the empty canvas shell so the surface and controls can be verified.
 */
export function StackCanvasTab(_props: StackCanvasTabProps) {
  return (
    <div className="h-[calc(100vh-13rem)] overflow-hidden rounded-md border border-border">
      <ReactFlowProvider>
        <CanvasEditor nodes={[]} edges={[]} />
      </ReactFlowProvider>
    </div>
  );
}
