import { Panel, useReactFlow } from "@xyflow/react";
import { Maximize2, Minus, Plus, Wand2, Workflow } from "lucide-react";
import { cn } from "@/lib/utils";
import { FIT_OPTIONS } from "./fit-options";

interface CanvasControlsProps {
  showConnections: boolean;
  onToggleConnections: () => void;
  onAutoLayout: () => void;
}

/**
 * Bottom-left control cluster: a zoom pill (in / out / fit) over a separate
 * connections-toggle square. Replaces React Flow's default `<Controls>` so the
 * chrome matches the design (and themes correctly in dark mode). Zoom actions
 * drive the pane via `useReactFlow`; must render inside a `ReactFlowProvider`.
 */
export function CanvasControls({ showConnections, onToggleConnections, onAutoLayout }: CanvasControlsProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow();

  const cell =
    "flex h-[38px] w-10 items-center justify-center text-muted-foreground transition-colors hover:text-brand";
  const square =
    "flex size-10 items-center justify-center rounded-md border bg-popover shadow-lg transition-colors";

  return (
    <Panel position="bottom-left" className="!m-4 flex flex-col gap-2">
      <div className="flex flex-col overflow-hidden rounded-md border border-border bg-popover shadow-lg">
        <button type="button" aria-label="Zoom in" className={cell} onClick={() => zoomIn()}>
          <Plus className="size-[17px]" />
        </button>
        <span className="h-px w-full bg-border" aria-hidden />
        <button type="button" aria-label="Zoom out" className={cell} onClick={() => zoomOut()}>
          <Minus className="size-[17px]" />
        </button>
        <span className="h-px w-full bg-border" aria-hidden />
        <button type="button" aria-label="Fit to view" title="Fit to view" className={cell} onClick={() => fitView(FIT_OPTIONS)}>
          <Maximize2 className="size-4" />
        </button>
      </div>
      <button
        type="button"
        aria-label="Auto layout"
        title="Auto layout"
        onClick={onAutoLayout}
        className={cn(square, "border-border text-muted-foreground hover:border-brand hover:text-brand")}
      >
        <Wand2 className="size-[18px]" />
      </button>
      <button
        type="button"
        aria-label={showConnections ? "Hide connections" : "Show connections"}
        aria-pressed={showConnections}
        title={showConnections ? "Hide connections" : "Show connections"}
        onClick={onToggleConnections}
        className={cn(
          square,
          showConnections ? "border-brand text-brand" : "border-border text-muted-foreground hover:text-foreground",
        )}
      >
        <Workflow className="size-[18px]" />
      </button>
    </Panel>
  );
}
