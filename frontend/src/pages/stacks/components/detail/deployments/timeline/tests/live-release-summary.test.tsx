// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "r18", sequence: 18, outcome: { resources: {} }, snapshot: { resources: [{ name: "web", source: { image: { ref: "nginx:1.25" } } }] } }) }));
import { LiveReleaseSummary } from "../live-release-summary";
import type { Stack } from "@/api/stacks";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const release = {
  id: "r18", sequence: 18, state: "Released", cause: { kind: "manual" },
  live_status: { resources: { web: { state: "Ready" } } },
} as unknown as StackRelease;
const stack = {
  current_release: { id: "r18" },
  spec: { stack_resources: [{ name: "web" }] },
} as unknown as Stack;
const ctx = { orgId: "o", teamName: "t", stackId: "s" };

describe("LiveReleaseSummary", () => {
  it("shows a lean live row, collapsed by default", () => {
    render(<LiveReleaseSummary release={release} stack={stack} logContext={ctx} />);
    expect(screen.getByRole("button", { name: /Live release #18/ })).toBeInTheDocument();
    expect(screen.getByText("Live")).toBeInTheDocument();
    expect(screen.getByText("#18")).toBeInTheDocument();
    // collapsed — the body (tracker / resource outcome) is not mounted yet
    expect(screen.queryByText("Resource outcome")).not.toBeInTheDocument();
  });

  it("expands in place into the live body (tracker + resource outcome)", async () => {
    render(<LiveReleaseSummary release={release} stack={stack} logContext={ctx} />);
    fireEvent.click(screen.getByRole("button", { name: /Live release #18/ }));
    expect(await screen.findByText("Resource outcome")).toBeInTheDocument();
    expect(screen.getByText("Build")).toBeInTheDocument();
    expect(screen.getAllByText("web").length).toBeGreaterThan(0);
  });
});
