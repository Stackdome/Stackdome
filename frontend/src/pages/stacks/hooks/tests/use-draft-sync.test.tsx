// @vitest-environment jsdom
import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { useStackEditSession } from "../use-stack-edit-session";
import { useDraftSync } from "../use-draft-sync";
import { DEBOUNCE_IDLE_MS, DEBOUNCE_MAX_WAIT_MS, SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";
import type { Stack } from "@/api/stacks";

vi.mock("@/api/stacks", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/stacks")>()),
  getStackById: vi.fn(),
}));
vi.mock("@/api/stack-resources", () => ({
  createStackResource: vi.fn(),
  updateStackResource: vi.fn(),
  deleteStackResource: vi.fn(),
}));
vi.mock("@/api/connections", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/connections")>()),
  createStackConnection: vi.fn(),
  updateStackConnection: vi.fn(),
  deleteStackConnection: vi.fn(),
}));
vi.mock("@/api/volumes", () => ({ createStackVolume: vi.fn(), deleteVolume: vi.fn() }));

import { getStackById } from "@/api/stacks";
import { updateStackResource, createStackResource } from "@/api/stack-resources";

const serverStack = (image: string): Stack =>
  ({
    id: "st-1",
    name: "demo",
    spec: {
      stack_resources: [{ id: "r-1", name: "web", source: { image: { ref: image } } }],
      volumes: [],
      connections: [],
    },
  }) as unknown as Stack;

const webForm = (image: string) => ({
  name: "web",
  sourceType: "image" as const,
  source: { image: { ref: image } },
});

function setup(stack: Stack) {
  const onStackRefreshed = vi.fn();
  const hook = renderHook(() => {
    const session = useStackEditSession();
    const sync = useDraftSync({
      enabled: true,
      stack,
      session,
      ids: { orgId: "o", teamName: "t", stackId: "st-1" },
      onStackRefreshed,
    });
    return { session, sync };
  });
  act(() => hook.result.current.session.start({ resources: [webForm("nginx:1")], volumes: [] }));
  return { hook, onStackRefreshed };
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.mocked(getStackById).mockResolvedValue(serverStack("nginx:2"));
  vi.mocked(updateStackResource).mockResolvedValue({} as never);
  vi.mocked(createStackResource).mockResolvedValue({} as never);
});
afterEach(() => {
  vi.useRealTimers();
  vi.clearAllMocks();
});

describe("useDraftSync", () => {
  it("debounces an edit and syncs after the idle window", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS - 100));
    expect(updateStackResource).not.toHaveBeenCalled();
    // Advance past the idle window; advanceTimersByTimeAsync flushes microtasks
    // too (the full async chain inside startCycle completes here).
    await act(() => vi.advanceTimersByTimeAsync(200));
    expect(updateStackResource).toHaveBeenCalledOnce();
    expect(updateStackResource).toHaveBeenCalledWith(
      "o",
      "t",
      "st-1",
      "web",
      expect.objectContaining({ name: "web" }),
    );
    // One more microtask flush so React applies the batched state updates.
    await act(() => Promise.resolve());
    expect(hook.result.current.sync.status).toBe(SYNC_STATUS.saved);
    // The session baseline stays pinned across autosaves — the edit must remain
    // visibly dirty (revertable) until a deploy or discard moves the baseline.
    expect(hook.result.current.session.dirty.dirtyResourceIdx.size).toBe(1);
  });

  it("coalesces rapid edits into one cycle", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    await act(() => vi.advanceTimersByTimeAsync(800));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:3")]));
    await act(() => vi.advanceTimersByTimeAsync(800));
    expect(updateStackResource).not.toHaveBeenCalled();
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS));
    expect(updateStackResource).toHaveBeenCalledOnce();
  });

  it("fires at max-wait under continuous edits", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    for (let i = 0; i < 8; i++) {
      act(() => hook.result.current.session.updateResources(() => [webForm(`nginx:${i + 2}`)]));
      await act(() => vi.advanceTimersByTimeAsync(700)); // always inside the idle window
      if (700 * (i + 1) > DEBOUNCE_MAX_WAIT_MS + 100) break;
    }
    expect(updateStackResource).toHaveBeenCalled();
  });

  it("skips API calls when the diff is structurally empty, but still rebases", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() =>
      hook.result.current.session.updateResources((prev) => [{ ...prev[0], depends_on: [] }]),
    );
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS + 100));
    expect(updateStackResource).not.toHaveBeenCalled();
    expect(createStackResource).not.toHaveBeenCalled();
  });

  it("on failure keeps the draft, reports error status, retries with backoff, recovers", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    vi.mocked(updateStackResource).mockRejectedValueOnce(new Error("500"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    // Advance past idle window; the cycle runs, fails, sets error status.
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS + 100));
    await act(() => Promise.resolve()); // flush batched React state
    expect(hook.result.current.sync.status).toBe(SYNC_STATUS.error);
    expect(hook.result.current.session.draft.resources[0].source?.image?.ref).toBe("nginx:2");
    // first backoff = RETRY_BASE_MS (1000ms); advance 1100ms to trigger retry
    await act(() => vi.advanceTimersByTimeAsync(1100));
    await act(() => Promise.resolve()); // flush batched React state
    expect(hook.result.current.sync.status).toBe(SYNC_STATUS.saved);
    expect(hook.result.current.sync.failureCount).toBe(0);
  });

  it("flush drains pending work and resolves true", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await hook.result.current.sync.flush();
    });
    expect(ok).toBe(true);
    expect(updateStackResource).toHaveBeenCalledOnce();
  });

  it("flush resolves false when the cycle fails", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    vi.mocked(updateStackResource).mockRejectedValue(new Error("500"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    let ok: boolean | undefined;
    await act(async () => {
      ok = await hook.result.current.sync.flush();
    });
    expect(ok).toBe(false);
  });

  it("does nothing when disabled", async () => {
    const onStackRefreshed = vi.fn();
    const hook = renderHook(() => {
      const session = useStackEditSession();
      const sync = useDraftSync({
        enabled: false,
        stack: serverStack("nginx:1"),
        session,
        ids: null,
        onStackRefreshed,
      });
      return { session, sync };
    });
    act(() => hook.result.current.session.start({ resources: [webForm("nginx:1")], volumes: [] }));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_MAX_WAIT_MS * 2));
    expect(updateStackResource).not.toHaveBeenCalled();
  });
});
