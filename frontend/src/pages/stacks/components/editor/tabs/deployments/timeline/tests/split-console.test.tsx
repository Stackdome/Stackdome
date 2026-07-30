// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
import { SplitConsole, type ResourceRowVM } from "../split-console";
import { ReleaseEventType, type ReleaseEvent } from "@/api/releases";
import { ResourceFailureType, compactEventMessage } from "../../derive";

afterEach(cleanup);

const ev = (over: Partial<ReleaseEvent> = {}): ReleaseEvent => ({
  sequence: over.sequence ?? 1,
  type: "release_started",
  message: "Deploying release",
  level: "info",
  occurred_at: "2026-07-19T03:43:48Z",
  ...over,
} as ReleaseEvent);

const rows: ResourceRowVM[] = [
  { name: "redis", phase: "Ready", source: { kind: "image", label: "redis:7" } },
  {
    name: "worker",
    phase: "CrashLoopBackOff",
    source: { kind: "image", label: "tooljet/tooljet:v3" },
    failure: { name: "worker", type: ResourceFailureType.Runtime, stage: "runtime", reason: "CrashLoopBackOff", message: "no match for platform in manifest", restartCount: 3 },
  },
];

const events: ReleaseEvent[] = [
  ev({ sequence: 1, message: "Release created" }),
  ev({ sequence: 2, type: ReleaseEventType.ResourceReady, resource_name: "redis", message: "redis is ready", level: "success" }),
  ev({ sequence: 3, type: ReleaseEventType.ResourceFailed, resource_name: "worker", message: "worker failed to start: image pull failed", level: "error" }),
];

describe("SplitConsole", () => {
  it("renders nothing with no rows, no events, not streaming", () => {
    const { container } = render(<SplitConsole rows={[]} events={[]} streaming={false} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("shows the rail with ready count and all events scoped to all resources", () => {
    render(<SplitConsole rows={rows} events={events} streaming={false} />);
    expect(screen.getByText("1/2 ready")).toBeInTheDocument();
    expect(screen.getByText("· all resources")).toBeInTheDocument();
    // Release-scoped event carries the "release" column tag.
    expect(screen.getByText("release")).toBeInTheDocument();
    expect(screen.getByText("Release created")).toBeInTheDocument();
    expect(screen.getByText("Failed to start — image pull failed")).toBeInTheDocument();
  });

  it("shows the live pulse only while streaming", () => {
    render(<SplitConsole rows={rows} events={events} streaming />);
    expect(screen.getByText("live")).toBeInTheDocument();
    cleanup();
    render(<SplitConsole rows={rows} events={events} streaming={false} />);
    expect(screen.queryByText("live")).not.toBeInTheDocument();
  });

  it("clicking a resource filters the console and pins its detail", async () => {
    render(<SplitConsole rows={rows} events={events} streaming={false} />);
    await userEvent.click(screen.getByRole("button", { name: /worker/ }));
    expect(screen.getByText("· worker")).toBeInTheDocument();
    // Pinned detail: status (also on the rail pill), source, failure message, restart count.
    expect(screen.getAllByText("CrashLoopBackOff").length).toBeGreaterThan(1);
    expect(screen.getByText("▢ tooljet/tooljet:v3")).toBeInTheDocument();
    expect(screen.getByText("no match for platform in manifest")).toBeInTheDocument();
    expect(screen.getByText("3 restarts")).toBeInTheDocument();
    // Console filtered to worker's events only.
    expect(screen.queryByText("Release created")).not.toBeInTheDocument();
    expect(screen.getByText("Failed to start — image pull failed")).toBeInTheDocument();
  });

  it("clicking all resources zooms back out", async () => {
    render(<SplitConsole rows={rows} events={events} streaming={false} />);
    await userEvent.click(screen.getByRole("button", { name: /worker/ }));
    await userEvent.click(screen.getByRole("button", { name: /all resources/ }));
    expect(screen.getByText("· all resources")).toBeInTheDocument();
    expect(screen.getByText("Release created")).toBeInTheDocument();
  });

  it("shows an empty message when the scope has no events", async () => {
    render(<SplitConsole rows={rows} events={[ev({ sequence: 1, message: "Release created" })]} streaming />);
    await userEvent.click(screen.getByRole("button", { name: /redis/ }));
    expect(screen.getByText("No activity yet")).toBeInTheDocument();
  });
});

describe("compactEventMessage", () => {
  it("compacts resource-scoped verbs and keeps the detail", () => {
    expect(compactEventMessage(ev({ type: ReleaseEventType.ResourceDeploying, message: "Deploying redis: StackResourceDeploymentNotReady" }))).toBe("Deploying — StackResourceDeploymentNotReady");
    expect(compactEventMessage(ev({ type: ReleaseEventType.ResourceWaiting, message: "worker is waiting: Dependent resource 'postgresql' not yet available" }))).toBe("Waiting — Dependent resource 'postgresql' not yet available");
    expect(compactEventMessage(ev({ type: ReleaseEventType.ResourceReady, message: "redis is ready" }))).toBe("Ready");
    expect(compactEventMessage(ev({ type: ReleaseEventType.ResourceFailed, message: "worker failed to start: image pull failed" }))).toBe("Failed to start — image pull failed");
  });

  it("keeps the verb alone when there is no detail and passes unknown types through", () => {
    expect(compactEventMessage(ev({ type: ReleaseEventType.ResourceDeploying, message: "Deploying redis" }))).toBe("Deploying");
    expect(compactEventMessage(ev({ message: "Release checks passed" }))).toBe("Release checks passed");
  });
});
