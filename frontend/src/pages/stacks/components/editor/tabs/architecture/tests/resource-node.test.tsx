// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, cleanup } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { ResourceNode } from "../nodes/resource-node";
import type { StatusVariant } from "@/components/branded/status-variant";

afterEach(cleanup);

function nodeProps(dotVariant: StatusVariant) {
  return {
    id: "n1", type: "resource", selected: false, zIndex: 0, isConnectable: false,
    positionAbsoluteX: 0, positionAbsoluteY: 0, dragging: false,
    data: {
      kind: "service", name: "web", kindLabel: "Web", glyph: "web",
      dotVariant, summary: "nginx", volumes: [],
    },
  } as never;
}

const dotOf = (container: HTMLElement) => container.querySelector("span[aria-hidden]");

describe("ResourceNode status dot", () => {
  it("ready dot breathes", () => {
    const { container } = render(<ReactFlowProvider><ResourceNode {...nodeProps("ready")} /></ReactFlowProvider>);
    expect(dotOf(container)!.className).toContain("animate-breathe");
    expect(dotOf(container)!.className).toContain("bg-success");
  });
  it("error dot is static red", () => {
    const { container } = render(<ReactFlowProvider><ResourceNode {...nodeProps("error")} /></ReactFlowProvider>);
    expect(dotOf(container)!.className).toContain("bg-danger");
    expect(dotOf(container)!.className).not.toContain("animate-breathe");
  });
  it("neutral dot is static gray", () => {
    const { container } = render(<ReactFlowProvider><ResourceNode {...nodeProps("neutral")} /></ReactFlowProvider>);
    expect(dotOf(container)!.className).toContain("bg-fg-muted");
  });
});
