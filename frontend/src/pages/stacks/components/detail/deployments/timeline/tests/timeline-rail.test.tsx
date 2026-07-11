// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({
  getRelease: vi.fn().mockResolvedValue({ id: "x", sequence: 1, outcome: { resources: {} }, snapshot: { resources: [] } }),
  listReleaseEvents: vi.fn().mockResolvedValue({ items: [] }),
  buildReleaseEventStreamUrl: vi.fn(() => ""),
  ReleaseEventScope: { Release: "release", Resource: "resource" },
}));
import { TimelineRail } from "../timeline-rail";
import { ReleaseDetailProvider } from "../../use-release-detail";
import type { ReleaseDetail } from "../../use-release-detail";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const stack = { spec: { stack_resources: [] } } as unknown as Stack;
const rels = (n: number): StackRelease[] => Array.from({ length: n }, (_, i) => ({ id: `r${n - i}`, sequence: n - i, state: "Released", cause: { kind: "manual" } } as StackRelease));

const base = { stack, onRollback: vi.fn(), onCancel: vi.fn(), onCopyId: vi.fn() };

const stubDetail: ReleaseDetail = { ensure: vi.fn(), peek: () => ({ loading: false }), refresh: vi.fn() };
function renderRail(ui: React.ReactElement) {
  return render(<ReleaseDetailProvider value={stubDetail}>{ui}</ReleaseDetailProvider>);
}

describe("TimelineRail", () => {
  it("renders the empty state with no releases", () => {
    renderRail(<TimelineRail releases={[]} {...base} />);
    expect(screen.getByText("No deployments yet")).toBeInTheDocument();
  });

  it("renders one continuous list with no Current/Earlier headers", () => {
    const r = rels(3);
    renderRail(<TimelineRail releases={r} activeRelease={r[0]} {...base} />);
    expect(screen.queryByText("Current deployment")).not.toBeInTheDocument();
    expect(screen.queryByText("Earlier deployments")).not.toBeInTheDocument();
    expect(screen.getByText("#3")).toBeInTheDocument();
    expect(screen.getByText("#2")).toBeInTheDocument();
    expect(screen.getByText("#1")).toBeInTheDocument();
  });

  it("opens the latest deploy by default and tags the live release", () => {
    const r = rels(3);
    const liveStack = { current_release: { id: "r2" }, spec: { stack_resources: [] } } as unknown as Stack;
    renderRail(<TimelineRail releases={r} activeRelease={r[0]} {...base} stack={liveStack} />);
    // #2 is the live release → carries the LIVE chip.
    expect(screen.getByText("Live")).toBeInTheDocument();
  });

  it("renders only the live dot solid; every other dot is a hollow ring", () => {
    const r = rels(3);
    const liveStack = { current_release: { id: "r2" }, spec: { stack_resources: [] } } as unknown as Stack;
    renderRail(<TimelineRail releases={r} activeRelease={r[0]} {...base} stack={liveStack} />);
    const dots = screen.getAllByTestId("rail-dot");
    const solid = dots.filter((d) => !d.className.includes("border-2") && !d.className.includes("animate-spin"));
    const ring = dots.filter((d) => d.className.includes("border-2"));
    expect(solid).toHaveLength(1); // only the live release (#2) is filled
    expect(ring.length).toBe(2); // #3 and #1 are hollow rings
  });

  it("windows earlier releases behind Show more", async () => {
    const r = rels(20);
    renderRail(<TimelineRail releases={r} activeRelease={r[0]} initialWindow={5} {...base} />);
    expect(screen.queryByText("#1")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /show more/i }));
    expect(screen.getByText("#1")).toBeInTheDocument();
  });

  it("auto-opens a release that first appears after mount (deploy while the tab is open)", () => {
    const r1 = { id: "r1", sequence: 1, state: "Released", cause: { kind: "manual" } } as StackRelease;
    const r2 = { id: "r2", sequence: 2, state: "Released", cause: { kind: "manual" } } as StackRelease;
    const r3 = { id: "r3", sequence: 3, state: "Released", cause: { kind: "manual" } } as StackRelease;
    const { container, rerender } = renderRail(<TimelineRail releases={[r2, r1]} activeRelease={r2} {...base} />);
    // r3 doesn't exist at mount, so the open-set initializer can't have opened it.
    expect(container.querySelector("#deploy-node-r3")).toBeNull();

    rerender(
      <ReleaseDetailProvider value={stubDetail}>
        <TimelineRail releases={[r3, r2, r1]} activeRelease={r3} {...base} />
      </ReleaseDetailProvider>,
    );
    // The new release's node is expanded (chevron rotated) — its live progress isn't hidden in a collapsed row.
    expect(container.querySelector("#deploy-node-r3")?.querySelector(".rotate-180")).toBeTruthy();
  });

  it("keeps multiple release details open at once (not an accordion)", async () => {
    const r = rels(4); // active #4, earlier #3 #2 #1
    renderRail(<TimelineRail releases={r} activeRelease={r[0]} {...base} />);
    await userEvent.click(screen.getByText("#3"));
    await userEvent.click(screen.getByText("#2"));
    // Both post-mortems stay mounted — an accordion would have closed #3 on opening #2.
    expect(await screen.findByText(/vs #2/i)).toBeInTheDocument();
    expect(await screen.findByText(/vs #1/i)).toBeInTheDocument();
  });
});
