// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";

vi.mock("@/api/image-builds", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/api/image-builds")>();
  return { ...actual, getImageBuild: vi.fn() };
});

import { getImageBuild, type ImageBuild } from "@/api/image-builds";
import { useBuildLogStream } from "../use-build-log-stream";

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  url: string;
  listeners = new Map<string, ((e: MessageEvent) => void)[]>();
  onopen: (() => void) | null = null;
  onmessage: ((e: MessageEvent) => void) | null = null;
  readyState = 0;
  closed = false;

  constructor(url: string) {
    this.url = url;
    FakeEventSource.instances.push(this);
  }

  addEventListener(type: string, cb: (e: MessageEvent) => void) {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), cb]);
  }

  close() {
    this.closed = true;
  }

  emit(type: string, data?: string) {
    if (type === "message") this.onmessage?.(new MessageEvent("message", { data }));
    else this.listeners.get(type)?.forEach((cb) => cb(new MessageEvent(type, { data })));
  }
}

const startedBuild = {
  id: "b1",
  status: { state: "Building", conditions: [{ type: "BuildJobCreated", status: "True" }] },
} as unknown as ImageBuild;

const props = { orgId: "o", projectName: "p", stackId: "s", buildId: "b1", enabled: true };

beforeEach(() => {
  FakeEventSource.instances = [];
  vi.stubGlobal("EventSource", FakeEventSource as unknown as typeof EventSource);
  vi.mocked(getImageBuild).mockReset();
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("useBuildLogStream", () => {
  it("opens the stream once the build job exists and accumulates lines", async () => {
    vi.mocked(getImageBuild).mockResolvedValue(startedBuild);

    const { result } = renderHook(() => useBuildLogStream(props));

    await waitFor(() => expect(result.current.phase).toBe("streaming"));
    expect(FakeEventSource.instances).toHaveLength(1);
    expect(FakeEventSource.instances[0].url).toContain("follow=true");
    expect(FakeEventSource.instances[0].url).toContain("tail=200");

    act(() => FakeEventSource.instances[0].emit("message", "INFO step 1"));
    act(() => FakeEventSource.instances[0].emit("message", "INFO step 2"));
    expect(result.current.lines).toEqual(["INFO step 1", "INFO step 2"]);
  });

  it("stays waiting while the build job is not created", async () => {
    vi.mocked(getImageBuild).mockResolvedValue({ status: {} } as unknown as ImageBuild);

    const { result } = renderHook(() => useBuildLogStream(props));

    await waitFor(() => expect(getImageBuild).toHaveBeenCalled());
    expect(result.current.phase).toBe("waiting");
    expect(FakeEventSource.instances).toHaveLength(0);
  });

  it("re-polls the build while it stays uncreated", async () => {
    vi.useFakeTimers();
    vi.mocked(getImageBuild)
      .mockResolvedValueOnce({ status: {} } as unknown as ImageBuild)
      .mockResolvedValue(startedBuild);

    const { result } = renderHook(() => useBuildLogStream(props));

    await act(() => vi.advanceTimersByTimeAsync(0));
    expect(getImageBuild).toHaveBeenCalledTimes(1);
    expect(result.current.phase).toBe("waiting");

    await act(() => vi.advanceTimersByTimeAsync(3000));

    expect(getImageBuild).toHaveBeenCalledTimes(2);
    expect(result.current.phase).toBe("streaming");
  });

  it("flips to ended and refetches the final build on the end event", async () => {
    const done = {
      ...startedBuild,
      status: { ...startedBuild.status, state: "Success" },
    } as unknown as ImageBuild;
    vi.mocked(getImageBuild).mockResolvedValueOnce(startedBuild).mockResolvedValueOnce(done);

    const { result } = renderHook(() => useBuildLogStream(props));

    await waitFor(() => expect(result.current.phase).toBe("streaming"));
    act(() => FakeEventSource.instances[0].emit("end"));

    await waitFor(() => expect(result.current.phase).toBe("ended"));
    expect(getImageBuild).toHaveBeenCalledTimes(2);
    await waitFor(() => expect(result.current.build).toEqual(done));
    expect(result.current.connectionStatus).toBe("disconnected");
    expect(FakeEventSource.instances[0].closed).toBe(true);
  });

  it("surfaces a backend stream error", async () => {
    vi.mocked(getImageBuild).mockResolvedValue(startedBuild);

    const { result } = renderHook(() => useBuildLogStream(props));

    await waitFor(() => expect(result.current.phase).toBe("streaming"));
    act(() => FakeEventSource.instances[0].emit("error", "pod pruned"));

    await waitFor(() => expect(result.current.phase).toBe("error"));
    expect(result.current.error).toBe("pod pruned");
  });

  it("marks the build unavailable when the preflight fetch fails", async () => {
    vi.mocked(getImageBuild).mockRejectedValue(new Error("not found"));

    const { result } = renderHook(() => useBuildLogStream(props));

    await waitFor(() => expect(result.current.phase).toBe("unavailable"));
    expect(result.current.error).toBe("not found");
    expect(FakeEventSource.instances).toHaveLength(0);
  });

  it("does nothing while disabled", () => {
    renderHook(() => useBuildLogStream({ ...props, enabled: false }));

    expect(getImageBuild).not.toHaveBeenCalled();
    expect(FakeEventSource.instances).toHaveLength(0);
  });

  it("stops polling after unmount", async () => {
    vi.useFakeTimers();
    vi.mocked(getImageBuild).mockResolvedValue({ status: {} } as unknown as ImageBuild);

    const { unmount } = renderHook(() => useBuildLogStream(props));

    await act(() => vi.advanceTimersByTimeAsync(0));
    expect(getImageBuild).toHaveBeenCalledTimes(1);

    unmount();
    await act(() => vi.advanceTimersByTimeAsync(9000));

    expect(getImageBuild).toHaveBeenCalledTimes(1);
    expect(FakeEventSource.instances).toHaveLength(0);
  });

  it("closes the stream when enabled flips to false", async () => {
    vi.mocked(getImageBuild).mockResolvedValue(startedBuild);

    const { result, rerender } = renderHook((p: typeof props) => useBuildLogStream(p), {
      initialProps: props,
    });

    await waitFor(() => expect(result.current.phase).toBe("streaming"));
    expect(FakeEventSource.instances[0].closed).toBe(false);

    rerender({ ...props, enabled: false });

    expect(FakeEventSource.instances[0].closed).toBe(true);
    expect(FakeEventSource.instances).toHaveLength(1);
  });

  it("surfaces a failed final-verdict refetch instead of rejecting unhandled", async () => {
    vi.mocked(getImageBuild)
      .mockResolvedValueOnce(startedBuild)
      .mockRejectedValueOnce(new Error("build record gone"));

    const { result } = renderHook(() => useBuildLogStream(props));

    await waitFor(() => expect(result.current.phase).toBe("streaming"));
    act(() => FakeEventSource.instances[0].emit("end"));

    await waitFor(() => expect(result.current.error).toBe("build record gone"));
    expect(result.current.phase).toBe("ended");
  });

  it("retry restarts the preflight", async () => {
    vi.mocked(getImageBuild).mockResolvedValue(startedBuild);

    const { result } = renderHook(() => useBuildLogStream(props));

    await waitFor(() => expect(result.current.phase).toBe("streaming"));
    act(() => result.current.retry());

    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(2));
    expect(FakeEventSource.instances[0].closed).toBe(true);
    expect(getImageBuild).toHaveBeenCalledTimes(2);
  });
});
