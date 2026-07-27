// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { useState } from "react";
import { ReactFlowProvider } from "@xyflow/react";
import { SidebarProvider } from "@/components/ui/sidebar";
import { HeaderCollapseContext } from "@/pages/stacks/lib/canvas/header-collapse";
import { CanvasControls } from "../canvas-controls";

afterEach(cleanup);

// SidebarProvider's mobile detection needs matchMedia, which jsdom lacks.
window.matchMedia ??= ((query: string) => ({
  matches: false,
  media: query,
  addEventListener: () => {},
  removeEventListener: () => {},
  addListener: () => {},
  removeListener: () => {},
  onchange: null,
  dispatchEvent: () => false,
})) as typeof window.matchMedia;

function Harness({ onAutoLayout = () => {} }: { onAutoLayout?: () => void }) {
  const [collapsed, setCollapsed] = useState(false);
  return (
    <SidebarProvider>
      <HeaderCollapseContext.Provider value={{ collapsed, setCollapsed }}>
        <ReactFlowProvider>
          <CanvasControls showConnections onToggleConnections={() => {}} onAutoLayout={onAutoLayout} />
        </ReactFlowProvider>
      </HeaderCollapseContext.Provider>
    </SidebarProvider>
  );
}

describe("CanvasControls", () => {
  it("groups auto layout with zen mode and keeps connections as the last control", () => {
    render(<Harness />);
    const zen = screen.getByRole("button", { name: "Zen mode" });
    const layout = screen.getByRole("button", { name: "Auto layout" });
    // Same pill: shared bordered parent.
    expect(layout.parentElement).toBe(zen.parentElement);
    // Connections toggle is the panel's last control.
    const panel = screen.getByRole("button", { name: "Hide connections" }).parentElement!;
    expect(panel.lastElementChild).toBe(screen.getByRole("button", { name: "Hide connections" }));
  });

  it("toggles zen mode via Cmd+.", () => {
    render(<Harness />);
    fireEvent.keyDown(window, { key: ".", metaKey: true });
    expect(screen.getByRole("button", { name: "Exit zen mode" })).toBeInTheDocument();
    fireEvent.keyDown(window, { key: ".", metaKey: true });
    expect(screen.getByRole("button", { name: "Zen mode" })).toBeInTheDocument();
  });

  it("zen button click collapses header and closes sidebar together", () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Zen mode" }));
    expect(screen.getByRole("button", { name: "Exit zen mode" })).toHaveAttribute("aria-pressed", "true");
  });

  it("auto layout button fires the callback", () => {
    const onAutoLayout = vi.fn();
    render(<Harness onAutoLayout={onAutoLayout} />);
    fireEvent.click(screen.getByRole("button", { name: "Auto layout" }));
    expect(onAutoLayout).toHaveBeenCalled();
  });
});
