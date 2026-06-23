// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
import { CurrentReleaseNode } from "../current-release-node";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);

const stack = (over: Record<string, unknown> = {}) => ({
  status: { resources: [{ name: "web", phase: "Ready", available_replicas: 1, replicas: 1 }], last_converged: { release_id: "r1" } },
  spec: { stack_resources: [{ name: "web", status: { state: "Ready" } }] },
  ...over,
}) as unknown as Stack;

function Wrap({ release, st, onCancel }: { release: StackRelease; st: Stack; onCancel?: (id: string) => void }) {
  return <CurrentReleaseNode release={release} stack={st} logContext={{ orgId: "o", teamName: "t", stackId: "s" }} onCancel={onCancel} />;
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

  it("does not render a changelog toggle (current node shows live status only)", () => {
    render(<Wrap release={{ id: "r1", sequence: 14, state: "Released" } as StackRelease} st={stack()} />);
    expect(screen.queryByRole("button", { name: /changelog/i })).not.toBeInTheDocument();
  });

  it("offers Cancel for a non-terminal release and fires onCancel", async () => {
    const onCancel = vi.fn();
    render(<Wrap release={{ id: "r9", sequence: 15, state: "Pending" } as StackRelease} st={stack()} onCancel={onCancel} />);
    await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
    expect(onCancel).toHaveBeenCalledWith("r9");
  });

  it("hides Cancel once the release is terminal", () => {
    render(<Wrap release={{ id: "r1", sequence: 14, state: "Released" } as StackRelease} st={stack()} onCancel={vi.fn()} />);
    expect(screen.queryByRole("button", { name: /cancel/i })).not.toBeInTheDocument();
  });
});
