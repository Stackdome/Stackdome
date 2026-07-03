// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 14, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { DeploymentsTab } from "../deployments-tab";
import type { DeployLifecycle } from "../use-deploy-lifecycle";
import type { Stack } from "@/api/stacks";
import type { StackRelease } from "@/api/releases";
import type { SnapshotDiff } from "../release-snapshot-diff";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const stack = { status: { resources: [] }, spec: { stack_resources: [] } } as unknown as Stack;
const releases: StackRelease[] = [{ id: "r1", sequence: 14, state: "Released", cause: { kind: "manual" } } as StackRelease];
const handlers = { onRollback: vi.fn(), onCancel: vi.fn(), onCopyId: vi.fn() };
const base = { orgId: "o", teamName: "t", stackId: "s", stack, loading: false, error: null, ...handlers };

const stagedDiff: SnapshotDiff = {
  resources: [{ name: "web-server", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image", from: "nginx:1.25", to: "nginx:1.27", kind: "changed" }] }] }],
  volumes: [],
  connections: [],
};

describe("DeploymentsTab", () => {
  it("renders the timeline with no in-tab Deploy button (deploy is owned by the status bar)", () => {
    const lifecycle: DeployLifecycle = { phase: "clean", nextSeq: 15, vsSeq: 14 };
    render(<DeploymentsTab {...base} releases={releases} activeRelease={releases[0]} lifecycle={lifecycle} />);
    expect(screen.getByText("#14")).toBeInTheDocument();
    expect(screen.getByText("Deploy timeline")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^deploy$/i })).not.toBeInTheDocument();
  });

  it("leads the rail with a draft node when staged", () => {
    const lifecycle: DeployLifecycle = { phase: "staged", stagedDiff, vsSeq: 14, nextSeq: 15 };
    render(<DeploymentsTab {...base} releases={releases} activeRelease={releases[0]} lifecycle={lifecycle} />);
    expect(screen.getByText("Draft")).toBeInTheDocument();
  });

  it("does not show a draft node when clean", () => {
    const lifecycle: DeployLifecycle = { phase: "clean", nextSeq: 15, vsSeq: 14 };
    render(<DeploymentsTab {...base} releases={releases} activeRelease={releases[0]} lifecycle={lifecycle} />);
    expect(screen.queryByText("Draft")).not.toBeInTheDocument();
  });

  it("renders an error state", () => {
    const lifecycle: DeployLifecycle = { phase: "clean", nextSeq: 1 };
    render(<DeploymentsTab {...base} error="boom" releases={[]} lifecycle={lifecycle} />);
    expect(screen.getByText("Could not load deployments")).toBeInTheDocument();
  });

  it("pins a live anchor when the live release is buried below a newer deploy", () => {
    const buriedStack = { status: { resources: [{ name: "web", phase: "Ready" }], last_converged: { release_id: "r14" } }, spec: { stack_resources: [{ name: "web" }] } } as unknown as Stack;
    const buried: StackRelease[] = [
      { id: "r15", sequence: 15, state: "InProgress", cause: { kind: "manual" } } as StackRelease,
      { id: "r14", sequence: 14, state: "Released", cause: { kind: "manual" } } as StackRelease,
    ];
    const lifecycle: DeployLifecycle = { phase: "deploying", nextSeq: 16 };
    render(<DeploymentsTab {...base} stack={buriedStack} releases={buried} activeRelease={buried[0]} lifecycle={lifecycle} />);
    expect(screen.getByRole("button", { name: /Live release #14/ })).toBeInTheDocument();
  });

  it("does not pin a live anchor when the live release is already the newest node", () => {
    const liveTopStack = { status: { resources: [], last_converged: { release_id: "r14" } }, spec: { stack_resources: [] } } as unknown as Stack;
    const lifecycle: DeployLifecycle = { phase: "clean", nextSeq: 15, vsSeq: 14 };
    render(<DeploymentsTab {...base} stack={liveTopStack} releases={releases} activeRelease={releases[0]} lifecycle={lifecycle} />);
    expect(screen.queryByRole("button", { name: /Live release/ })).not.toBeInTheDocument();
  });
});
