// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("../../build-logs-modal", () => ({
  BuildLogsModal: ({ buildId, resourceName }: { buildId: string; resourceName: string }) => (
    <div data-testid="build-logs-modal">{buildId}:{resourceName}</div>
  ),
}));
import { SplitConsole, type ResourceRowVM } from "../split-console";
import { fetchLogSnapshot } from "@/api/observability";
import { BuildLogsLinkTarget, ReleaseEventLinkKind, ReleaseEventType, type ReleaseEvent } from "@/api/releases";
import { ResourceFailureType, compactEventMessage } from "../../derive";

afterEach(() => {
  cleanup();
  vi.mocked(fetchLogSnapshot).mockReset().mockResolvedValue([]);
});

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

const logContext = { orgId: "o-1", projectName: "proj", stackId: "s-1" };

const buildLogsLink = (target: Record<string, string>) => ({
  kind: ReleaseEventLinkKind.BuildLogs,
  label: "View build logs",
  target,
});

const buildLogsEvent = (over: Partial<ReleaseEvent> = {}): ReleaseEvent => ev({
  sequence: 10,
  type: ReleaseEventType.BuildStarted,
  resource_name: "api",
  message: "Building api",
  links: [buildLogsLink({ [BuildLogsLinkTarget.BuildID]: "b-1", [BuildLogsLinkTarget.ResourceName]: "api" })],
  ...over,
});

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
    expect(screen.getByText("Failed to start — CrashLoopBackOff")).toBeInTheDocument();
  });

  it("badges the all-resources row with the resource count, not the event count", () => {
    render(<SplitConsole rows={rows} events={events} streaming={false} />);
    const allBtn = screen.getByRole("button", { name: /all resources/ });
    expect(allBtn).toHaveTextContent("2");
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
    expect(screen.getByText("Failed to start — CrashLoopBackOff")).toBeInTheDocument();
  });

  it("shows an OOM diagnosis first and keeps captured termination output out of the activity line", async () => {
    const terminationOutput = "allocation failed\nprocess shutting down";
    const oomRows: ResourceRowVM[] = [{
      name: "tooljet",
      phase: "CrashLoopBackOff",
      failure: {
        name: "tooljet",
        type: ResourceFailureType.Runtime,
        stage: "runtime",
        failureType: "out_of_memory",
        reason: "OOMKilled",
        message: terminationOutput,
        exitCode: 137,
        restartCount: 1,
      },
    }];
    const oomEvents = [ev({
      type: ReleaseEventType.ResourceFailed,
      resource_name: "tooljet",
      message: `tooljet failed to start: ${terminationOutput}`,
      level: "error",
    })];

    render(<SplitConsole rows={oomRows} events={oomEvents} streaming={false} logContext={logContext} />);
    expect(screen.getByText("Failed to start — Out of memory (OOMKilled), exit 137")).not.toHaveTextContent("allocation failed");

    await userEvent.click(screen.getByRole("button", { name: /tooljet/ }));
    expect(screen.getByText("Out of memory (OOMKilled)")).toBeInTheDocument();
    expect(screen.getByText("exit 137")).toBeInTheDocument();
    expect(screen.getByText("1 restart")).toBeInTheDocument();
    expect(screen.getByText("Termination output")).toBeInTheDocument();
    expect(screen.getByText((_, element) => element?.tagName === "PRE" && element.textContent === terminationOutput)).toBeInTheDocument();
  });

  it("fetches the crash-log snapshot only when captured termination output is absent", async () => {
    vi.mocked(fetchLogSnapshot).mockResolvedValueOnce(["fallback crash log"]);
    const crashRows: ResourceRowVM[] = [{
      name: "worker",
      phase: "CrashLoopBackOff",
      failure: {
        name: "worker",
        type: ResourceFailureType.Runtime,
        stage: "runtime",
        failureType: "crash_loop",
        reason: "CrashLoopBackOff",
      },
    }];

    render(<SplitConsole rows={crashRows} events={[]} streaming logContext={logContext} />);
    await userEvent.click(screen.getByRole("button", { name: /worker/ }));

    expect(await screen.findByText("fallback crash log")).toBeInTheDocument();
    expect(screen.queryByText("Termination output")).not.toBeInTheDocument();
  });

  it("uses captured output as the primary fallback when structured fields are absent", async () => {
    const messageRows: ResourceRowVM[] = [{
      name: "worker",
      phase: "Error",
      failure: {
        name: "worker",
        type: ResourceFailureType.Runtime,
        stage: "runtime",
        reason: undefined,
        message: "captured termination output",
      },
    }];

    render(<SplitConsole rows={messageRows} events={[]} streaming logContext={logContext} />);
    await userEvent.click(screen.getByRole("button", { name: /worker/ }));

    expect(screen.getByText("captured termination output")).toBeInTheDocument();
    expect(screen.queryByText("Termination output")).not.toBeInTheDocument();
  });

  it("keeps historical failure events tied to their own recorded reason", () => {
    const currentRows: ResourceRowVM[] = [{
      name: "worker",
      phase: "CrashLoopBackOff",
      failure: {
        name: "worker",
        type: ResourceFailureType.Runtime,
        stage: "runtime",
        failureType: "out_of_memory",
        reason: "OOMKilled",
        exitCode: 137,
      },
    }];
    const failureEvents = [
      ev({
        sequence: 1,
        type: ReleaseEventType.ResourceFailed,
        resource_name: "worker",
        message: "worker failed to start: old image pull log tail",
        metadata: { reason: "ImagePullBackOff" },
        level: "error",
      }),
      ev({
        sequence: 2,
        type: ReleaseEventType.ResourceFailed,
        resource_name: "worker",
        message: "worker failed to start: current OOM log tail",
        metadata: { reason: "OOMKilled" },
        level: "error",
      }),
    ];

    render(<SplitConsole rows={currentRows} events={failureEvents} streaming={false} />);

    expect(screen.getByText("Failed to start — ImagePullBackOff")).toBeInTheDocument();
    expect(screen.getByText("Failed to start — Out of memory (OOMKilled), exit 137")).toBeInTheDocument();
    expect(screen.queryByText(/old image pull log tail/)).not.toBeInTheDocument();
    expect(screen.queryByText(/current OOM log tail/)).not.toBeInTheDocument();
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

  it("opens the build logs modal from a build_logs link", async () => {
    render(<SplitConsole rows={[]} events={[buildLogsEvent()]} streaming logContext={logContext} />);
    expect(screen.queryByTestId("build-logs-modal")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /View build logs/ }));
    expect(screen.getByTestId("build-logs-modal")).toHaveTextContent("b-1:api");
  });

  it("falls back to the event's resource for the modal when the link target omits it", async () => {
    const event = buildLogsEvent({ links: [buildLogsLink({ [BuildLogsLinkTarget.BuildID]: "b-1" })] });
    render(<SplitConsole rows={[]} events={[event]} streaming logContext={logContext} />);
    await userEvent.click(screen.getByRole("button", { name: /View build logs/ }));
    expect(screen.getByTestId("build-logs-modal")).toHaveTextContent("b-1:api");
  });

  it("keeps build_logs links inert without a log context", () => {
    render(<SplitConsole rows={[]} events={[buildLogsEvent()]} streaming />);
    expect(screen.queryByRole("button", { name: /View build logs/ })).not.toBeInTheDocument();
    expect(screen.getByText(/View build logs/)).toBeInTheDocument();
  });

  it("renders links of other kinds as labels", () => {
    const event = buildLogsEvent({ links: [{ kind: "some_future_kind", label: "Other link" }] });
    render(<SplitConsole rows={[]} events={[event]} streaming logContext={logContext} />);
    expect(screen.queryByRole("button", { name: /Other link/ })).not.toBeInTheDocument();
    expect(screen.getByText(/Other link/)).toBeInTheDocument();
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
