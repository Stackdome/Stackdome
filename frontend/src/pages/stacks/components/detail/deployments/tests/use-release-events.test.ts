// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";

vi.mock("@/api/releases", () => ({
  listReleaseEvents: vi.fn(),
  buildReleaseEventStreamUrl: vi.fn(
    (orgId: string, teamName: string, stackId: string, releaseId: string, afterSequence?: number) =>
      `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/releases/${releaseId}/events/stream${
        afterSequence !== undefined ? `?after_sequence=${afterSequence}` : ""
      }`,
  ),
}));

import { listReleaseEvents, type ReleaseEvent } from "@/api/releases";
import { useReleaseEvents } from "../use-release-events";

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  url: string;
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onerror: (() => void) | null = null;
  closed = false;

  constructor(url: string) {
    this.url = url;
    FakeEventSource.instances.push(this);
  }

  open() { this.onopen?.(); }
  emit(data: ReleaseEvent) { this.onmessage?.({ data: JSON.stringify(data) }); }
  error() { this.onerror?.(); }
  close() { this.closed = true; }
}

const ev = (sequence: number, over: Partial<ReleaseEvent> = {}): ReleaseEvent => ({ id: `e${sequence}`, sequence, message: `m${sequence}`, ...over });

beforeEach(() => {
  FakeEventSource.instances = [];
  vi.stubGlobal("EventSource", FakeEventSource as unknown as typeof EventSource);
  vi.mocked(listReleaseEvents).mockReset();
  vi.mocked(listReleaseEvents).mockResolvedValue({ items: [] });
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("useReleaseEvents", () => {
  it("idle when releaseId is undefined — no fetch, no stream", () => {
    const { result } = renderHook(() => useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", terminal: false }));
    expect(result.current.status).toBe("idle");
    expect(result.current.events).toEqual([]);
    expect(FakeEventSource.instances).toHaveLength(0);
    expect(listReleaseEvents).not.toHaveBeenCalled();
  });

  it("streams, dedupes, and orders events by sequence", async () => {
    const onEvent = vi.fn();
    const { result } = renderHook(() =>
      useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", releaseId: "r1", terminal: false, onEvent }),
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));
    const source = FakeEventSource.instances[0];

    act(() => source.open());
    expect(result.current.status).toBe("streaming");

    act(() => {
      source.emit(ev(1));
      source.emit(ev(3));
      source.emit(ev(2));
      source.emit(ev(1)); // duplicate, ignored
    });

    expect(result.current.events.map((e) => e.sequence)).toEqual([1, 2, 3]);
    expect(onEvent).toHaveBeenCalledTimes(3);
  });

  it("reconnects after an error, resuming from the last seen sequence", async () => {
    vi.useFakeTimers();
    renderHook(() => useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", releaseId: "r1", terminal: false }));
    expect(FakeEventSource.instances).toHaveLength(1);
    const first = FakeEventSource.instances[0];
    act(() => {
      first.open();
      first.emit(ev(5));
    });

    act(() => first.error());
    expect(first.closed).toBe(true);

    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });

    expect(FakeEventSource.instances).toHaveLength(2);
    expect(FakeEventSource.instances[1].url).toContain("after_sequence=5");
  });

  it("terminal release: one-shot list fetch, no EventSource, ends closed", async () => {
    vi.mocked(listReleaseEvents).mockResolvedValue({ items: [ev(1), ev(2)] });
    const { result } = renderHook(() =>
      useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", releaseId: "r1", terminal: true }),
    );

    await waitFor(() => expect(result.current.status).toBe("closed"));
    expect(result.current.events.map((e) => e.sequence)).toEqual([1, 2]);
    expect(FakeEventSource.instances).toHaveLength(0);
    expect(listReleaseEvents).toHaveBeenCalledTimes(1);
    expect(listReleaseEvents).toHaveBeenCalledWith("o", "t", "s", "r1");
  });

  it("closes the stream on unmount", async () => {
    const { unmount } = renderHook(() =>
      useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", releaseId: "r1", terminal: false }),
    );
    expect(FakeEventSource.instances).toHaveLength(1);
    const source = FakeEventSource.instances[0];
    unmount();
    expect(source.closed).toBe(true);
  });

  it("keeps prior events visible across a non-terminal→terminal transition (no flash of empty)", async () => {
    let resolveList!: (v: { items: ReleaseEvent[] }) => void;
    vi.mocked(listReleaseEvents).mockImplementation(
      () => new Promise((res) => { resolveList = res; }),
    );

    const { result, rerender } = renderHook(
      ({ terminal }) => useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", releaseId: "r1", terminal }),
      { initialProps: { terminal: false } },
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));
    const source = FakeEventSource.instances[0];
    act(() => {
      source.open();
      source.emit(ev(1));
      source.emit(ev(2));
    });
    expect(result.current.events.map((e) => e.sequence)).toEqual([1, 2]);

    // Deploy completes: terminal flips true for the SAME releaseId.
    rerender({ terminal: true });
    // Still showing the streamed events while the one-shot terminal fetch is in flight.
    expect(result.current.events.map((e) => e.sequence)).toEqual([1, 2]);

    await act(async () => { resolveList({ items: [ev(1), ev(2), ev(3)] }); });
    await waitFor(() => expect(result.current.status).toBe("closed"));
    expect(result.current.events.map((e) => e.sequence)).toEqual([1, 2, 3]);
  });

  it("resets events when the releaseId itself changes", async () => {
    const { result, rerender } = renderHook(
      ({ releaseId }) => useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", releaseId, terminal: false }),
      { initialProps: { releaseId: "r1" } },
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));
    act(() => {
      FakeEventSource.instances[0].open();
      FakeEventSource.instances[0].emit(ev(1));
    });
    expect(result.current.events.map((e) => e.sequence)).toEqual([1]);

    rerender({ releaseId: "r2" });
    expect(result.current.events).toEqual([]);
  });

  it("falls back to polling once MAX_RECONNECTS is exceeded", async () => {
    vi.useFakeTimers();
    const { result } = renderHook(() =>
      useReleaseEvents({ orgId: "o", teamName: "t", stackId: "s", releaseId: "r1", terminal: false }),
    );

    for (let i = 0; i < 11; i++) {
      const current = FakeEventSource.instances[FakeEventSource.instances.length - 1];
      act(() => current.error());
      await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    }

    expect(result.current.status).toBe("polling");

    vi.mocked(listReleaseEvents).mockResolvedValueOnce({ items: [ev(9)] });
    await act(async () => { await vi.advanceTimersByTimeAsync(5000); });
    expect(result.current.events.map((e) => e.sequence)).toEqual([9]);
    expect(listReleaseEvents).toHaveBeenCalledWith("o", "t", "s", "r1", 0);
  });
});
