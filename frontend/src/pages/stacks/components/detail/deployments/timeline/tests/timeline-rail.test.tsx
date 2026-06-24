// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "x", sequence: 1, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { TimelineRail } from "../timeline-rail";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const stack = { status: { resources: [] }, spec: { stack_resources: [] } } as unknown as Stack;
const rels = (n: number): StackRelease[] => Array.from({ length: n }, (_, i) => ({ id: `r${n - i}`, sequence: n - i, state: "Released", cause: { kind: "manual" } } as StackRelease));

const base = { stack, onRollback: vi.fn(), onCancel: vi.fn(), onCopyId: vi.fn() };

describe("TimelineRail", () => {
  it("renders the empty state with no releases", () => {
    render(<TimelineRail releases={[]} {...base} />);
    expect(screen.getByText("No deployments yet")).toBeInTheDocument();
  });

  it("renders current node + earlier releases", () => {
    const r = rels(3);
    render(<TimelineRail releases={r} activeRelease={r[0]} {...base} />);
    expect(screen.getByText("Current deployment")).toBeInTheDocument();
    expect(screen.getByText("Earlier deployments")).toBeInTheDocument();
    expect(screen.getByText("#2")).toBeInTheDocument();
    expect(screen.getByText("#1")).toBeInTheDocument();
  });

  it("windows earlier releases behind Show more", async () => {
    const r = rels(20);
    render(<TimelineRail releases={r} activeRelease={r[0]} initialWindow={5} {...base} />);
    expect(screen.queryByText("#1")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /show more/i }));
    expect(screen.getByText("#1")).toBeInTheDocument();
  });

  it("keeps multiple release details open at once (not an accordion)", async () => {
    const r = rels(4); // active #4, earlier #3 #2 #1
    render(<TimelineRail releases={r} activeRelease={r[0]} {...base} />);
    await userEvent.click(screen.getByText("#3"));
    await userEvent.click(screen.getByText("#2"));
    // Both post-mortems stay mounted — an accordion would have closed #3 on opening #2.
    expect(await screen.findByText(/vs #2/i)).toBeInTheDocument();
    expect(await screen.findByText(/vs #1/i)).toBeInTheDocument();
  });
});
