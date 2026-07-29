// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
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

describe("ResourceNode port lines", () => {
  const details = [
    { text: "port 80 · public", port: 80, public: true },
    { text: "port 89 · public", port: 89, public: true },
    { text: "port 9090 · internal", port: 9090, public: false },
  ];

  function withPorts(portUrls?: Record<number, string>) {
    const p = nodeProps("ready") as { data: Record<string, unknown> };
    p.data = { ...p.data, details, portUrls };
    return p as never;
  }

  it("renders one compact line of port numbers; public+deployed link out", () => {
    render(
      <ReactFlowProvider>
        <ResourceNode {...withPorts({ 80: "https://web.acme.stackdome.app" })} />
      </ReactFlowProvider>,
    );
    expect(screen.getByText("ports")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: "80" });
    expect(link).toHaveAttribute("href", "https://web.acme.stackdome.app");
    expect(link).toHaveAttribute("target", "_blank");
    // port 89 has no live URL yet; internal ports never link.
    expect(screen.getByText("89")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "89" })).toBeNull();
    expect(screen.queryByRole("link", { name: "9090" })).toBeNull();
    // full description survives as the tooltip
    expect(screen.getByText("89")).toHaveAttribute("title", "port 89 · public");
  });

  it("renders plain numbers without portUrls", () => {
    render(<ReactFlowProvider><ResourceNode {...withPorts(undefined)} /></ReactFlowProvider>);
    expect(screen.getByText("80")).toBeInTheDocument();
    expect(screen.queryByRole("link")).toBeNull();
  });
});

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
