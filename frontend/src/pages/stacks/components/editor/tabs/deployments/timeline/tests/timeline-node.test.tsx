// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({
  getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 13, outcome: { resources: {} }, snapshot: { resources: [] } }),
  listReleaseEvents: vi.fn().mockResolvedValue({ items: [] }),
  buildReleaseEventStreamUrl: vi.fn(() => ""),
  ReleaseEventScope: { Release: "release", Resource: "resource" },
  ReleaseEventType: { ResourceWaiting: "resource_waiting", ResourceDeploying: "resource_deploying", ResourceReady: "resource_ready", ResourceFailed: "resource_failed" },
}));
import { useReleaseDetail } from "../../use-release-detail";
import { ConfirmProvider } from "@/components/branded/confirm";
import { TimelineNode } from "../timeline-node";
import { getRelease, listReleaseEvents, ReleaseEventScope } from "@/api/releases";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onerror: (() => void) | null = null;
  constructor(public url: string) { FakeEventSource.instances.push(this); }
  open() { this.onopen?.(); }
  emit(data: unknown) { this.onmessage?.({ data: JSON.stringify(data) }); }
  close() { /* noop */ }
}

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});
beforeEach(() => {
  FakeEventSource.instances = [];
  vi.stubGlobal("EventSource", FakeEventSource as unknown as typeof EventSource);
});
afterEach(() => vi.unstubAllGlobals());

const stack = { spec: { stack_resources: [] } } as unknown as Stack;

function Wrap(over: Partial<React.ComponentProps<typeof TimelineNode>> & { release: StackRelease }) {
  const detail = useReleaseDetail("o", "t", "s");
  return (
    <TimelineNode
      detail={detail}
      isOpen={false}
      onToggle={vi.fn()}
      onRollback={vi.fn()}
      onCancel={vi.fn()}
      isActive={false}
      isLive={false}
      stack={stack}
      {...over}
    />
  );
}

describe("TimelineNode", () => {
  it("shows the sequence + cause and toggles on click", async () => {
    const onToggle = vi.fn();
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released", cause: { kind: "rollback", detail: "9" } } as StackRelease} onToggle={onToggle} />);
    expect(screen.getByText("#13")).toBeInTheDocument();
    expect(screen.getByText("Rollback to #9")).toBeInTheDocument();
    await userEvent.click(screen.getByText("#13"));
    expect(onToggle).toHaveBeenCalledWith("r1");
  });

  it("shows a Failed release's message in danger", () => {
    render(<Wrap release={{ id: "r1", sequence: 9, state: "Failed", message: "apply quota error" } as StackRelease} />);
    expect(screen.getByText(/apply quota error/)).toHaveClass("text-danger");
  });

  it("tags the live release with a LIVE chip", () => {
    render(<Wrap release={{ id: "r1", sequence: 7, state: "Released" } as StackRelease} isLive />);
    expect(screen.getByText("Live")).toBeInTheDocument();
  });

  it("renders the live body (stage tracker) for the active node when open", () => {
    render(<Wrap release={{ id: "r1", sequence: 7, state: "Released" } as StackRelease} isActive isOpen />);
    expect(screen.getByText("Build")).toBeInTheDocument();
    expect(screen.getByText("Deploy")).toBeInTheDocument();
  });

  it("offers rollback on an earlier release, and only after confirmation", async () => {
    const onRollback = vi.fn();
    render(
      <ConfirmProvider>
        <Wrap release={{ id: "r1", sequence: 12, state: "Released" } as StackRelease} isOpen onRollback={onRollback} />
      </ConfirmProvider>,
    );
    await userEvent.click(screen.getByRole("button", { name: /Rollback to this/ }));
    expect(onRollback).not.toHaveBeenCalled();
    await userEvent.click(await screen.findByRole("button", { name: /^Roll back$/ }));
    await waitFor(() => expect(onRollback).toHaveBeenCalledWith("r1"));
  });

  it("offers no rollback on the live release — it is already serving traffic", () => {
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released" } as StackRelease} isOpen isLive />);
    expect(screen.queryByRole("button", { name: /Rollback to this/ })).not.toBeInTheDocument();
  });

  it("renders the historical post-mortem for a non-active node when open", async () => {
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released" } as StackRelease} isOpen prevReleaseId="r0" prevSeq={12} />);
    await userEvent.click(await screen.findByRole("button", { name: /Changes/ }));
    expect(screen.getByText(/vs #12/i)).toBeInTheDocument();
  });

  it("renders async validation errors on the live body and jumps to the offending resource", async () => {
    // validation_errors ride the DETAIL payload; list items carry none.
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: "r1",
      sequence: 9,
      state: "Failed",
      snapshot: { resources: [] },
      validation_errors: [{ resource_name: "web", field: "source.image.ref", code: "image_not_found", message: "not found" }],
    });
    const onJumpToResource = vi.fn();
    const stackWithResource = { spec: { stack_resources: [{ name: "web" }] } } as unknown as Stack;
    render(
      <Wrap
        release={{ id: "r1", sequence: 9, state: "Failed" } as StackRelease}
        isActive
        isOpen
        stack={stackWithResource}
        onJumpToResource={onJumpToResource}
      />,
    );
    await userEvent.click(await screen.findByText(/image_not_found/));
    expect(onJumpToResource).toHaveBeenCalledWith("web", "configuration");
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r1", sequence: 13, outcome: { resources: {} }, snapshot: { resources: [] } });
  });

  it("renders release activity events fetched for a terminal release's live body", async () => {
    (listReleaseEvents as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      items: [{ id: "e1", sequence: 1, type: "build_started", level: "info", message: "Building web", resource_name: "web" }],
    });
    render(<Wrap release={{ id: "r1", sequence: 7, state: "Released" } as StackRelease} isActive isOpen />);
    expect(await screen.findByText("Building web")).toBeInTheDocument();
  });

  it("hybrid driver: a release-scoped event refetches releases and refreshes the release detail", async () => {
    const refetchReleases = vi.fn();
    render(
      <Wrap
        release={{ id: "r1", sequence: 9, state: "InProgress" } as StackRelease}
        isActive
        isOpen
        refetchReleases={refetchReleases}
      />,
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));
    const source = FakeEventSource.instances[0];
    const getReleaseCallsBefore = (getRelease as ReturnType<typeof vi.fn>).mock.calls.length;
    act(() => {
      source.open();
      source.emit({ id: "e1", sequence: 1, scope: ReleaseEventScope.Release, type: "release_released", message: "Released" });
    });
    expect(refetchReleases).toHaveBeenCalledTimes(1);
    await waitFor(() => expect((getRelease as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(getReleaseCallsBefore));
  });
});
