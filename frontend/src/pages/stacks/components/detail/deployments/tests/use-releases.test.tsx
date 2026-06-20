// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, cleanup } from "@testing-library/react";

vi.mock("@/api/releases", () => ({ listReleases: vi.fn() }));
import { listReleases } from "@/api/releases";
import { useReleases } from "../use-releases";

const ARGS = { orgId: "o", teamName: "t", stackId: "s", enabled: true };

beforeEach(() => vi.clearAllMocks());
afterEach(() => { cleanup(); vi.useRealTimers(); });

describe("useReleases", () => {
  it("fetches on mount and exposes the active release (highest sequence first)", async () => {
    (listReleases as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [{ id: "r2", sequence: 2, state: "Released" }, { id: "r1", sequence: 1, state: "Superseded" }],
      total: 2,
    });
    const { result } = renderHook(() => useReleases(ARGS));
    await waitFor(() => expect(result.current.releases).toHaveLength(2));
    expect(result.current.activeRelease?.id).toBe("r2");
    expect(listReleases).toHaveBeenCalledWith("o", "t", "s");
  });

  it("does not fetch when disabled", async () => {
    renderHook(() => useReleases({ ...ARGS, enabled: false }));
    await Promise.resolve();
    expect(listReleases).not.toHaveBeenCalled();
  });

  it("polls while the active release is non-terminal", async () => {
    vi.useFakeTimers();
    (listReleases as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [{ id: "r1", sequence: 1, state: "InProgress" }] });
    renderHook(() => useReleases(ARGS));
    await vi.advanceTimersByTimeAsync(0);
    expect(listReleases).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(5000);
    expect(listReleases).toHaveBeenCalledTimes(2);
  });
});
