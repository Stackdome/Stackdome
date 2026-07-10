// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 13, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { useReleaseDetail } from "../../use-release-detail";
import { TimelineNode } from "../timeline-node";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

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
      onCopyId={vi.fn()}
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

  it("renders the historical post-mortem for a non-active node when open", async () => {
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released" } as StackRelease} isOpen prevReleaseId="r0" prevSeq={12} />);
    expect(await screen.findByText(/vs #12/i)).toBeInTheDocument();
  });
});
