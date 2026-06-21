// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 14, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { useReleaseDetail } from "../../use-release-detail";
import { CurrentReleaseNode } from "../current-release-node";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);

const stack = (over: Record<string, unknown> = {}) => ({
  status: { resources: [{ name: "web", phase: "Ready", available_replicas: 1, replicas: 1 }], last_converged: { release_id: "r1" } },
  spec: { stack_resources: [{ name: "web", status: { state: "Ready" } }] },
  ...over,
}) as unknown as Stack;

function Wrap({ release, st }: { release: StackRelease; st: Stack }) {
  const detail = useReleaseDetail("o", "t", "s");
  return <CurrentReleaseNode release={release} stack={st} detail={detail} logContext={{ orgId: "o", teamName: "t", stackId: "s" }} prevReleaseId="r0" prevSeq={13} />;
}

describe("CurrentReleaseNode", () => {
  it("renders status, sequence and a Ready resource row", () => {
    render(<Wrap release={{ id: "r1", sequence: 14, state: "Released", snapshot_revision: "abcdef1234" } as StackRelease} st={stack()} />);
    expect(screen.getByText("Released")).toBeInTheDocument();
    expect(screen.getByText("#14")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
  });

  it("renders the release-error block for a Failed release with no per-resource failure", () => {
    const st = stack({ status: { resources: [] }, spec: { stack_resources: [] } });
    render(<Wrap release={{ id: "r1", sequence: 17, state: "Failed", message: "apply error: unknown addon" } as StackRelease} st={st} />);
    expect(screen.getByText(/apply error: unknown addon/)).toBeInTheDocument();
  });

  it("toggles its own changelog", async () => {
    render(<Wrap release={{ id: "r1", sequence: 14, state: "Released" } as StackRelease} st={stack()} />);
    await userEvent.click(screen.getByRole("button", { name: /view changelog/i }));
    expect(await screen.findByText(/nothing to compare/i)).toBeInTheDocument();
  });
});
