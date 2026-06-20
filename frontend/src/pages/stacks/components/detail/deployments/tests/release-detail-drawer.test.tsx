// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";

vi.mock("@/api/releases", () => ({ getRelease: vi.fn() }));
import { getRelease } from "@/api/releases";
import { ReleaseDetailDrawer } from "../release-detail-drawer";

beforeAll(() => {
  Element.prototype.hasPointerCapture = () => false;
  Element.prototype.releasePointerCapture = () => {};
  Element.prototype.setPointerCapture = () => {};
  Element.prototype.scrollIntoView = () => {};
});
afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("ReleaseDetailDrawer", () => {
  it("loads the full release and shows outcomes + message", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: "r1", sequence: 14, state: "Failed", message: "timed out waiting for convergence after 15m0s",
      outcome: { resources: { tooljet: { phase: "Failed", ready_replicas: 0, replicas: 1, message: "CrashLoopBackOff" } } },
      pins: { resources: { tooljet: { git_sha: "9c69af2" } } },
      snapshot: { spec: { x: 1 } },
    });
    render(<ReleaseDetailDrawer orgId="o" teamName="t" stackId="s" releaseId="r1" onClose={vi.fn()} />);
    await waitFor(() => expect(screen.getByText(/timed out waiting for convergence/)).toBeInTheDocument());
    expect(screen.getByText("tooljet")).toBeInTheDocument();
    expect(getRelease).toHaveBeenCalledWith("o", "t", "s", "r1");
  });

  it("renders config changes vs the previous release snapshot", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: "r2", sequence: 15, state: "Released", snapshot: { spec: { replicas: 2 } },
    });
    render(<ReleaseDetailDrawer orgId="o" teamName="t" stackId="s" releaseId="r2"
      previousRelease={{ id: "r1", snapshot: { spec: { replicas: 1 } } } as never} onClose={vi.fn()} />);
    await waitFor(() => expect(screen.getByText(/Config changes/i)).toBeInTheDocument());
    expect(screen.getByText("spec.replicas")).toBeInTheDocument();
  });
});
