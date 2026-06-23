// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 13, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { useReleaseDetail } from "../../use-release-detail";
import { HistoryRow } from "../history-row";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

function Wrap({ release, isOpen, onToggle, prevReleaseId }: { release: StackRelease; isOpen: boolean; onToggle: (id: string) => void; prevReleaseId?: string }) {
  const detail = useReleaseDetail("o", "t", "s");
  return <HistoryRow release={release} detail={detail} isOpen={isOpen} onToggle={onToggle} onRollback={vi.fn()} onCancel={vi.fn()} onCopyId={vi.fn()} prevReleaseId={prevReleaseId} prevSeq={12} />;
}

describe("HistoryRow", () => {
  it("shows cause + state and toggles open on click", async () => {
    const onToggle = vi.fn();
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released", cause: { kind: "rollback", detail: "9" } } as StackRelease} isOpen={false} onToggle={onToggle} />);
    expect(screen.getByText("#13")).toBeInTheDocument();
    expect(screen.getByText("Rollback to #9")).toBeInTheDocument();
    await userEvent.click(screen.getByText("#13"));
    expect(onToggle).toHaveBeenCalledWith("r1");
  });

  it("renders the post-mortem when open", async () => {
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released" } as StackRelease} isOpen onToggle={vi.fn()} prevReleaseId="r0" />);
    // Predecessor present + identical snapshots → the no-changes copy (not "initial").
    expect(await screen.findByText(/no configuration changes since #12/i)).toBeInTheDocument();
  });

  it("shows the failure message in danger for a Failed release", () => {
    render(<Wrap release={{ id: "r1", sequence: 9, state: "Failed", message: "apply quota error" } as StackRelease} isOpen={false} onToggle={vi.fn()} />);
    expect(screen.getByText(/apply quota error/)).toHaveClass("text-danger");
  });
});
