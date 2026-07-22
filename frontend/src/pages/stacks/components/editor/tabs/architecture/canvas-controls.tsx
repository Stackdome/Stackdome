import { useContext } from "react";
import { Panel, useReactFlow } from "@xyflow/react";
import { Focus, Maximize2, Minus, Plus, Wand2, Workflow } from "lucide-react";
import { cn } from "@/lib/utils";
import { useSidebar } from "@/components/ui/sidebar";
import { HeaderCollapseContext } from "@/pages/stacks/lib/canvas/header-collapse";
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
  const { open: sidebarOpen, setOpen: setSidebarOpen } = useSidebar();
  const { collapsed: headerCollapsed, setCollapsed: setHeaderCollapsed } = useContext(HeaderCollapseContext);

  // Zen: header collapsed + sidebar closed, toggled as one.
  const zenActive = headerCollapsed && !sidebarOpen;
  const toggleZen = () => {
    const next = !zenActive;
    setHeaderCollapsed(next);
    setSidebarOpen(!next);
  };

  const cell =
    "flex h-8 w-8 items-center justify-center text-muted-foreground transition-colors hover:text-brand";
  const square =
    "flex size-8 items-center justify-center rounded-md border bg-popover shadow-lg transition-colors";

  return (
    <Panel position="bottom-left" className="!m-4 flex flex-col gap-2">
      <div className="flex flex-col overflow-hidden rounded-md border border-border bg-popover shadow-lg">
        <button type="button" aria-label="Zoom in" className={cell} onClick={() => zoomIn()}>
          <Plus className="size-3.5" />
        </button>
        <span className="h-px w-full bg-border" aria-hidden />
        <button type="button" aria-label="Zoom out" className={cell} onClick={() => zoomOut()}>
          <Minus className="size-3.5" />
        </button>
        <span className="h-px w-full bg-border" aria-hidden />
        <button type="button" aria-label="Fit to view" title="Fit to view" className={cell} onClick={() => fitView(FIT_OPTIONS)}>
          <Maximize2 className="size-3.5" />
        </button>
      </div>
      <button
        type="button"
        aria-label="Auto layout"
        title="Auto layout"
        onClick={onAutoLayout}
        className={cn(square, "border-border text-muted-foreground hover:border-brand hover:text-brand")}
      >
        <Wand2 className="size-4" />
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
        <Workflow className="size-4" />
      </button>
      <button
        type="button"
        aria-label={zenActive ? "Exit zen mode" : "Zen mode"}
        aria-pressed={zenActive}
        title={zenActive ? "Exit zen mode" : "Zen mode — collapse header and sidebar"}
        onClick={toggleZen}
        className={cn(
          square,
          zenActive ? "border-brand text-brand" : "border-border text-muted-foreground hover:text-foreground",
        )}
      >
        <Focus className="size-4" />
      </button>
    </Panel>
  );
}
