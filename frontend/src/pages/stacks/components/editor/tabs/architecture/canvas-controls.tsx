import { useCallback, useContext, useEffect } from "react";
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
 * Bottom-left control cluster: a zoom pill (in / out / fit), a layout pill
 * (auto layout / zen mode), then the connections toggle. Replaces React Flow's
 * default `<Controls>` so the chrome matches the design (and themes correctly
 * in dark mode). Zoom actions drive the pane via `useReactFlow`; must render
 * inside a `ReactFlowProvider`.
 */
export function CanvasControls({ showConnections, onToggleConnections, onAutoLayout }: CanvasControlsProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow();
  const { open: sidebarOpen, setOpen: setSidebarOpen } = useSidebar();
  const { collapsed: headerCollapsed, setCollapsed: setHeaderCollapsed } = useContext(HeaderCollapseContext);

  // Zen: header collapsed + sidebar closed, toggled as one.
  const zenActive = headerCollapsed && !sidebarOpen;
  const toggleZen = useCallback(() => {
    const next = !zenActive;
    setHeaderCollapsed(next);
    setSidebarOpen(!next);
  }, [zenActive, setHeaderCollapsed, setSidebarOpen]);

  // ⌘. toggles zen. Lives here (not the shell) because zen also needs the
  // sidebar; the canvas stays mounted across tabs, so the shortcut is global.
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.defaultPrevented) return; // consumed by a nested layer (dialog, drawer…)
      if (e.key === "." && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        toggleZen();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [toggleZen]);

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
      <div className="flex flex-col overflow-hidden rounded-md border border-border bg-popover shadow-lg">
        <button
          type="button"
          aria-label="Auto layout"
          title="Auto layout"
          onClick={onAutoLayout}
          className={cell}
        >
          <Wand2 className="size-4" />
        </button>
        <span className="h-px w-full bg-border" aria-hidden />
        <button
          type="button"
          aria-label={zenActive ? "Exit zen mode" : "Zen mode"}
          aria-pressed={zenActive}
          title={zenActive ? "Exit zen mode (⌘.)" : "Zen mode — collapse header and sidebar (⌘.)"}
          onClick={toggleZen}
          className={cn(cell, zenActive && "text-brand")}
        >
          <Focus className="size-4" />
        </button>
      </div>
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
    </Panel>
  );
}
