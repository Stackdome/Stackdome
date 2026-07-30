// @vitest-environment jsdom
// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { CanvasContextMenu, type CanvasMenuTarget } from "../canvas-context-menu";

afterEach(cleanup);

const handlers = () => ({
  onClose: vi.fn(),
  onOpenResource: vi.fn(),
  onAddVolumeToResource: vi.fn(),
  onDeleteResource: vi.fn(),
  onDisconnectVolume: vi.fn(),
  onOpenVolume: vi.fn(),
  onRequestDeleteVolume: vi.fn(),
  onRequestAttach: vi.fn(),
});

const renderMenu = (target: CanvasMenuTarget) => {
  const h = handlers();
  render(<CanvasContextMenu target={target} {...h} />);
  return h;
};

describe("CanvasContextMenu", () => {
  it("shows chip items and fires disconnect", () => {
    const h = renderMenu({ kind: "volume-chip", volumeName: "data", x: 10, y: 10 });
    expect(screen.getByText("Disconnect volume")).toBeInTheDocument();
    expect(screen.getByText("Volume settings")).toBeInTheDocument();
    expect(screen.getByText("Delete volume")).toBeInTheDocument();
    expect(screen.queryByText("Attach to service…")).not.toBeInTheDocument();
    fireEvent.click(screen.getByText("Disconnect volume"));
    expect(h.onDisconnectVolume).toHaveBeenCalledWith("data");
  });

  it("shows floating-volume items and fires attach", () => {
    const h = renderMenu({ kind: "volume-node", volumeName: "data", x: 10, y: 10 });
    expect(screen.getByText("Attach to service…")).toBeInTheDocument();
    expect(screen.queryByText("Disconnect volume")).not.toBeInTheDocument();
    fireEvent.click(screen.getByText("Attach to service…"));
    expect(h.onRequestAttach).toHaveBeenCalledWith("data");
  });

  it("shows resource items and fires delete", () => {
    const h = renderMenu({ kind: "resource", resourceIdx: 2, resourceName: "web", x: 10, y: 10 });
    expect(screen.getByText("Open settings")).toBeInTheDocument();
    expect(screen.getByText("Add volume…")).toBeInTheDocument();
    fireEvent.click(screen.getByText("Delete service"));
    expect(h.onDeleteResource).toHaveBeenCalledWith("web");
  });
});
