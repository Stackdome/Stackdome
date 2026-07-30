// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useReleaseAnchors } from "@/pages/stacks/components/editor/hooks/use-release-anchors";
import { ReleaseState } from "@/pages/stacks/components/editor/tabs/deployments/release-states";
import type { Stack } from "@/pages/stacks/types";

const fakeDetail = () => ({
  ensure: vi.fn(),
  refresh: vi.fn(),
  peek: vi.fn(() => ({ data: undefined })),
}) as never; // structural stand-in for ReturnType<typeof useReleaseDetail>

const stackWith = (convergedReleaseId?: string, state: string = ReleaseState.Released) =>
  (convergedReleaseId
    ? ({ converged_release: { id: convergedReleaseId, state } } as unknown as Stack)
    : ({} as Stack));

describe("useReleaseAnchors", () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it("pins the baseline to the active release, falling back to converged_release", () => {
    const detail = fakeDetail();
    const { result } = renderHook(() =>
      useReleaseAnchors({
        stack: stackWith("cur-1"),
        releases: [],
        activeRelease: { id: "act-1", state: ReleaseState.Released },
        releaseDetail: detail,
      }),
    );
    expect(result.current.baselineReleaseId).toBe("act-1");
    expect((detail as { ensure: ReturnType<typeof vi.fn> }).ensure).toHaveBeenCalledWith("act-1");
  });

  it("falls back to converged_release as baseline before the releases list loads", () => {
    const { result } = renderHook(() =>
      useReleaseAnchors({ stack: stackWith("cur-1"), releases: [], activeRelease: undefined, releaseDetail: fakeDetail() }),
    );
    expect(result.current.baselineReleaseId).toBe("cur-1");
  });

  it("selects the non-terminal release as the status anchor during a rollout", () => {
    const { result } = renderHook(() =>
      useReleaseAnchors({
        stack: stackWith("cur-1"),
        releases: [
          { id: "done-1", state: ReleaseState.Released },
          { id: "rolling-1", state: ReleaseState.InProgress }, // real non-terminal state from release-states.ts
        ],
        activeRelease: { id: "done-1", state: ReleaseState.Released },
        releaseDetail: fakeDetail(),
      }),
    );
    expect(result.current.statusReleaseId).toBe("rolling-1");
  });

  it("polls the status release every 5s while non-terminal, and stops when terminal", () => {
    const detail = fakeDetail();
    const { unmount } = renderHook(() =>
      useReleaseAnchors({
        stack: stackWith("cur-1"),
        releases: [{ id: "rolling-1", state: ReleaseState.InProgress }],
        activeRelease: undefined,
        releaseDetail: detail,
      }),
    );
    const refresh = (detail as { refresh: ReturnType<typeof vi.fn> }).refresh;
    const callsAfterMount = refresh.mock.calls.length; // immediate refresh on mount
    vi.advanceTimersByTime(5000);
    expect(refresh.mock.calls.length).toBe(callsAfterMount + 1);
    unmount();
    vi.advanceTimersByTime(10000);
    expect(refresh.mock.calls.length).toBe(callsAfterMount + 1); // cleanup stopped the interval
  });

  it("does not poll when the status release is terminal", () => {
    const detail = fakeDetail();
    renderHook(() =>
      useReleaseAnchors({ stack: stackWith("cur-1", ReleaseState.Released), releases: [], activeRelease: undefined, releaseDetail: detail }),
    );
    const refresh = (detail as { refresh: ReturnType<typeof vi.fn> }).refresh;
    const calls = refresh.mock.calls.length;
    vi.advanceTimersByTime(15000);
    expect(refresh.mock.calls.length).toBe(calls);
  });
});
