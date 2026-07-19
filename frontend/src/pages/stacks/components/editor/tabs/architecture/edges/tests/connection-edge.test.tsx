// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, cleanup } from "@testing-library/react";
import { ConnectionEdge } from "../connection-edge";

afterEach(cleanup);

// The edge reads node rects from React Flow's store via useInternalNode; feed it
// fake internal nodes instead of mounting a full ReactFlow instance.
const nodes = new Map<string, unknown>();
vi.mock("@xyflow/react", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@xyflow/react")>();
  return {
    ...actual,
    useInternalNode: (id: string) => nodes.get(id),
    BaseEdge: ({ path }: { path: string }) => <path data-testid="edge-path" d={path} />,
  };
});

function setNode(id: string, x: number, y: number, width = 216, height = 104) {
  nodes.set(id, { internals: { positionAbsolute: { x, y } }, measured: { width, height } });
}

function renderEdge() {
  const props = { id: "e1", source: "a", target: "b" } as Parameters<typeof ConnectionEdge>[0];
  return render(
    <svg>
      <ConnectionEdge {...props} />
    </svg>,
  );
}

function arrowTransform(container: HTMLElement): string {
  const polygon = container.querySelector("polygon");
  expect(polygon).not.toBeNull();
  return polygon!.getAttribute("transform") ?? "";
}

describe("ConnectionEdge", () => {
  it("renders nothing while either endpoint node is missing", () => {
    nodes.clear();
    setNode("a", 0, 0);
    const { container } = renderEdge();
    expect(container.querySelector("polygon")).toBeNull();
    expect(container.querySelector('[data-testid="edge-path"]')).toBeNull();
  });

  it.each([
    ["below (enters bottom face)", 0, 500, "rotate(-90)"],
    ["above (enters top face)", 0, -500, "rotate(90)"],
    ["left of target (enters left face)", -800, 0, "rotate(0)"],
    ["right of target (enters right face)", 800, 0, "rotate(180)"],
  ])("points the arrowhead into the target when the source sits %s", (_desc, sx, sy, rotation) => {
    nodes.clear();
    setNode("a", sx, sy);
    setNode("b", 0, 0);
    const { container } = renderEdge();
    expect(arrowTransform(container)).toContain(rotation);
    expect(container.querySelector('[data-testid="edge-path"]')).not.toBeNull();
  });
});
