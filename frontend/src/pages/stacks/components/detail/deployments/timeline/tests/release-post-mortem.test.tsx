// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
vi.mock("@/api/releases", () => ({ getRelease: vi.fn() }));
import { getRelease } from "@/api/releases";
import { useReleaseDetail } from "../../use-release-detail";
import { ReleasePostMortem } from "../release-post-mortem";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);

const stack = { status: {}, spec: { stack_resources: [] } } as unknown as import("@/api/stacks").Stack;

function Wrap({ release, prevId }: { release: StackRelease; prevId?: string }) {
  const detail = useReleaseDetail("o", "t", "s");
  return <ReleasePostMortem detail={detail} release={release} stack={stack} prevReleaseId={prevId} prevSeq={12} />;
}

describe("ReleasePostMortem", () => {
  it("shows outcomes + config diff once loaded", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockImplementation((_o, _t, _s, id) =>
      Promise.resolve(id === "r-cur"
        ? { id, sequence: 13, outcome: { resources: { web: { phase: "Ready", ready_replicas: 1, replicas: 1 } } }, snapshot: { resources: [{ name: "web", image_spec: { image: "web:2" } }] } }
        : { id, sequence: 12, snapshot: { resources: [{ name: "web", image_spec: { image: "web:1" } }] } }));
    render(<Wrap release={{ id: "r-cur", sequence: 13, state: "Released" } as StackRelease} prevId="r-prev" />);
    await waitFor(() => expect(screen.getAllByText("web").length).toBeGreaterThan(0));
    expect(screen.getByText("Resource outcome")).toBeInTheDocument();
    // Build→Deploy→Ready tracker leads the card (uniform with the live body).
    expect(screen.getByText("Build")).toBeInTheDocument();
    // The image/repo source is sourced from the release snapshot (uniform with the live body).
    expect(screen.getByText("web:2")).toBeInTheDocument();
    // Config changes is a brand-colored collapsed toggle, not an always-open block.
    const toggle = screen.getByText(/Config changes · vs #12/);
    expect(toggle).toHaveClass("text-brand");
  });

  it("shows the red Deploy-failed banner for a failed release (uniform with the live body)", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r-cur", sequence: 9, outcome: { resources: {} }, snapshot: { resources: [] } });
    render(<Wrap release={{ id: "r-cur", sequence: 9, state: "Failed", message: "apply error: quota" } as StackRelease} />);
    await waitFor(() => expect(screen.getByText(/apply error: quota/)).toBeInTheDocument());
    expect(screen.getByText("Deploy failed")).toBeInTheDocument();
    expect(screen.queryByText("Why it failed")).not.toBeInTheDocument();
  });

  it("shows an error line if the fetch fails", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("nope"));
    render(<Wrap release={{ id: "r-cur", sequence: 5, state: "Released" } as StackRelease} />);
    await waitFor(() => expect(screen.getByText(/nope/)).toBeInTheDocument());
  });
});
