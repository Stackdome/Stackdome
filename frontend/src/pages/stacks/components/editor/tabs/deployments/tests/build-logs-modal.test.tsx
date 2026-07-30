// @vitest-environment jsdom
import { describe, expect, it, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, cleanup } from "@testing-library/react";

vi.mock("react-lazylog", () => ({
  LazyLog: ({ text }: { text: string }) => <pre data-testid="lazylog">{text}</pre>,
}));
vi.mock("../use-build-log-stream", () => ({ useBuildLogStream: vi.fn() }));

import { BuildLogsModal, BuildPhase, deriveBuildLogsView } from "../build-logs-modal";
import { useBuildLogStream } from "../use-build-log-stream";
import type { ImageBuild } from "@/api/image-builds";

afterEach(cleanup);

type StreamResult = ReturnType<typeof useBuildLogStream>;

const retry = vi.fn();
const base: StreamResult = {
  lines: [],
  phase: "waiting",
  connectionStatus: "disconnected",
  build: null,
  error: null,
  retry,
};

function mockStream(overrides: Partial<StreamResult>) {
  vi.mocked(useBuildLogStream).mockReturnValue({ ...base, ...overrides });
}

function buildWith(state: string, revision?: string): ImageBuild {
  return { status: { state, build_source_revision: revision } } as unknown as ImageBuild;
}

const props = {
  open: true,
  onClose: vi.fn(),
  orgId: "o",
  projectName: "p",
  stackId: "s",
  buildId: "b",
  resourceName: "api",
};

describe("BuildLogsModal", () => {
  it("shows the waiting state while the build job has not been created", () => {
    mockStream({ phase: "waiting" });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/Waiting for the build to start/)).toBeInTheDocument();
  });

  it("renders streamed lines with the resource name and short revision", () => {
    mockStream({
      phase: "streaming",
      connectionStatus: "connected",
      lines: ["l1", "l2"],
      build: buildWith(BuildPhase.Pending, "abcdef1234567"),
    });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByTestId("lazylog")).toHaveTextContent("l1 l2");
    expect(screen.getByText(/Build logs — api/)).toBeInTheDocument();
    expect(screen.getByText("abcdef1")).toBeInTheDocument();
    expect(screen.getByText(BuildPhase.Pending)).toBeInTheDocument();
  });

  it("shows a connecting state while streaming before the first line", () => {
    mockStream({ phase: "streaming", connectionStatus: "connecting" });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/Connecting to the build/)).toBeInTheDocument();
  });

  it("shows the success banner after the stream ends", () => {
    mockStream({ phase: "ended", lines: ["done"], build: buildWith(BuildPhase.Success) });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/Build succeeded — log stream complete/)).toBeInTheDocument();
  });

  it("shows the failure banner when the build failed", () => {
    mockStream({ phase: "ended", lines: ["boom"], build: buildWith(BuildPhase.Failed) });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/Build failed — log stream complete/)).toBeInTheDocument();
  });

  it("shows the expired message when the preflight found no build", () => {
    mockStream({ phase: "unavailable", error: "not found" });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/no longer available/)).toBeInTheDocument();
    expect(
      screen.getByText(
        "Logs for completed builds are kept only for a short time. The build's outcome and failure details remain on this build.",
      ),
    ).toBeInTheDocument();
  });

  it("keeps infrastructure vocabulary out of the copy the user sees", () => {
    const infraTerms = /\b(job|jobs|pod|pods|ttl|kubernetes|k8s|namespace|container|cluster)\b/i;
    for (const stream of [
      { phase: "unavailable" as const },
      { phase: "waiting" as const },
      { phase: "waiting" as const, build: buildWith(BuildPhase.Failed) },
      { phase: "streaming" as const, connectionStatus: "connecting" as const },
      { phase: "ended" as const, build: buildWith(BuildPhase.Success) },
      { phase: "error" as const, build: buildWith(BuildPhase.Pending) },
    ]) {
      mockStream(stream);
      render(<BuildLogsModal {...props} />);
      // Radix portals the dialog outside the render container.
      expect(document.body.textContent ?? "").not.toMatch(infraTerms);
      cleanup();
    }
  });

  it("treats a dead stream on a terminal build as expired logs, not a connection fault", () => {
    mockStream({
      phase: "streaming",
      connectionStatus: "disconnected",
      build: buildWith(BuildPhase.Success),
    });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/no longer available/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /retry/i })).not.toBeInTheDocument();
  });

  it("offers a retry when the stream dies while the build is still running", () => {
    mockStream({
      phase: "streaming",
      connectionStatus: "disconnected",
      build: buildWith(BuildPhase.Pending),
    });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/Log stream interrupted/)).toBeInTheDocument();
    expect(screen.queryByText(/Connecting to the build/)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /retry/i }));
    expect(retry).toHaveBeenCalled();
  });

  it("offers a retry when a running build's stream dies after emitting lines", () => {
    mockStream({
      phase: "streaming",
      connectionStatus: "disconnected",
      lines: ["l1"],
      build: buildWith(BuildPhase.Pending),
    });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/Log stream interrupted/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });

  it("ends the wait when the build reached a terminal state before its job existed", () => {
    mockStream({ phase: "waiting", build: buildWith(BuildPhase.Failed) });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/ended before it produced logs/)).toBeInTheDocument();
    expect(screen.queryByText(/Waiting for the build to start/)).not.toBeInTheDocument();
  });

  it("offers a retry when the stream drops before the build finished", () => {
    mockStream({ phase: "error", error: "stream reset", build: buildWith(BuildPhase.Pending) });
    render(<BuildLogsModal {...props} />);
    expect(screen.getByText(/Log stream interrupted/)).toBeInTheDocument();
    expect(screen.getByText("stream reset")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /retry/i }));
    expect(retry).toHaveBeenCalled();
  });

  it("renders nothing when closed", () => {
    mockStream({ phase: "streaming", lines: ["l1"] });
    render(<BuildLogsModal {...props} open={false} />);
    expect(screen.queryByTestId("lazylog")).not.toBeInTheDocument();
  });
});

describe("deriveBuildLogsView", () => {
  const input = {
    phase: "streaming" as const,
    connectionStatus: "connected" as const,
    buildDone: false,
    hasLines: false,
  };

  it("prefers expiry over a retry prompt once the build is terminal", () => {
    expect(deriveBuildLogsView({ ...input, phase: "error", buildDone: true })).toBe("expired");
    expect(deriveBuildLogsView({ ...input, phase: "error", buildDone: false })).toBe("interrupted");
  });

  it("treats a dropped connection on a live build as interrupted, never as connecting", () => {
    const dropped = { ...input, connectionStatus: "disconnected" as const };
    expect(deriveBuildLogsView(dropped)).toBe("interrupted");
    expect(deriveBuildLogsView({ ...dropped, hasLines: true })).toBe("interrupted");
    expect(deriveBuildLogsView({ ...dropped, buildDone: true })).toBe("expired");
  });

  it("keeps waiting only while the build is still live", () => {
    expect(deriveBuildLogsView({ ...input, phase: "waiting" })).toBe("waiting");
    expect(deriveBuildLogsView({ ...input, phase: "waiting", buildDone: true })).toBe(
      "endedWithoutLogs",
    );
  });

  it("reports an empty completed stream separately from a live one", () => {
    expect(deriveBuildLogsView({ ...input, phase: "ended" })).toBe("noOutput");
    expect(deriveBuildLogsView({ ...input, phase: "ended", hasLines: true })).toBe("log");
    expect(deriveBuildLogsView({ ...input })).toBe("connecting");
  });
});
