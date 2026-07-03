// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { DrawerStack } from "../DrawerStack";
import type { DrawerEntry } from "@/pages/stacks/lib/canvas/drawer-stack";

afterEach(cleanup);

const r = (index: number): DrawerEntry => ({ kind: "resource", index });
const v = (name: string): DrawerEntry => ({ kind: "volume", name });

function setup(overrides: Partial<Parameters<typeof DrawerStack>[0]> = {}) {
  const onTruncate = vi.fn();
  const onPop = vi.fn();
  const onCloseAll = vi.fn();
  render(
    <DrawerStack
      panels={[
        { entry: r(0), title: "web", icon: <span /> },
        { entry: v("data"), title: "data", icon: <span /> },
      ]}
      front={<div>volume body</div>}
      onTruncate={onTruncate}
      onPop={onPop}
      onCloseAll={onCloseAll}
      {...overrides}
    />,
  );
  return { onTruncate, onPop, onCloseAll };
}

describe("DrawerStack", () => {
  it("renders the front panel body and a header-only behind panel", () => {
    setup();
    expect(screen.getByText("volume body")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Bring web panel to front" })).toBeInTheDocument();
  });

  it("staggers behind panels and layers z-indices", () => {
    setup();
    const behind = screen.getByTestId("drawer-panel-0");
    const front = screen.getByTestId("drawer-panel-1");
    expect(behind.style.right).toBe("28px"); // 12 + 16·1
    expect(behind.style.zIndex).toBe("199");
    expect(front.style.right).toBe("12px");
    expect(front.style.zIndex).toBe("200");
  });

  it("clicking a behind panel truncates to its depth", () => {
    const { onTruncate } = setup();
    fireEvent.click(screen.getByRole("button", { name: "Bring web panel to front" }));
    expect(onTruncate).toHaveBeenCalledWith(0);
  });

  it("Escape pops, Shift+Escape closes all", () => {
    const { onPop, onCloseAll } = setup();
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onPop).toHaveBeenCalledTimes(1);
    fireEvent.keyDown(window, { key: "Escape", shiftKey: true });
    expect(onCloseAll).toHaveBeenCalledTimes(1);
  });

  it("renders nothing when the stack is empty", () => {
    render(
      <DrawerStack panels={[]} front={null} onTruncate={vi.fn()} onPop={vi.fn()} onCloseAll={vi.fn()} />,
    );
    expect(screen.queryByTestId("drawer-panel-0")).toBeNull();
  });
});
