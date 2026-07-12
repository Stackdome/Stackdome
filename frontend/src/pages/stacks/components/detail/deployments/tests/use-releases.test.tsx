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

  it("does not fast-poll while a release is non-terminal — events drive updates instead", async () => {
    vi.useFakeTimers();
    (listReleases as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [{ id: "r1", sequence: 1, state: "InProgress" }] });
    renderHook(() => useReleases(ARGS));
    await vi.advanceTimersByTimeAsync(0);
    expect(listReleases).toHaveBeenCalledTimes(1);
    // No 5s fast poll left — only the 30s idle poll remains.
    await vi.advanceTimersByTimeAsync(5000);
    await vi.advanceTimersByTimeAsync(20000);
    expect(listReleases).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(5000);
    expect(listReleases).toHaveBeenCalledTimes(2);
  });

  it("coalesces a refetch requested while one is in flight (terminal event isn't dropped)", async () => {
    const mock = listReleases as ReturnType<typeof vi.fn>;
    let resolveFirst!: (v: unknown) => void;
    mock
      .mockImplementationOnce(() => new Promise((res) => { resolveFirst = res; }))
      .mockResolvedValue({ items: [{ id: "r1", sequence: 1, state: "Released" }] });

    const { result } = renderHook(() => useReleases(ARGS));
    await waitFor(() => expect(mock).toHaveBeenCalledTimes(1)); // mount fetch, still in flight

    // Event-driven refetch lands mid-flight — must queue, not vanish.
    result.current.refetch();
    expect(mock).toHaveBeenCalledTimes(1);

    resolveFirst({ items: [] });
    await waitFor(() => expect(mock).toHaveBeenCalledTimes(2)); // queued refetch runs after settle
    await waitFor(() => expect(result.current.releases).toHaveLength(1));
  });

  it("still does a slow idle poll when everything is terminal (catches external deploys)", async () => {
    vi.useFakeTimers();
    const mock = listReleases as ReturnType<typeof vi.fn>;
    mock.mockResolvedValue({ items: [{ id: "r1", sequence: 1, state: "Released" }] });
    renderHook(() => useReleases(ARGS));
    await vi.advanceTimersByTimeAsync(0);
    expect(mock).toHaveBeenCalledTimes(1);
    // fast poll skips (terminal); the 30s idle poll still refetches
    await vi.advanceTimersByTimeAsync(30000);
    expect(mock).toHaveBeenCalledTimes(2);
  });
});
