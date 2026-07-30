// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  attachSseHandlers,
  aggregateStreamStatus,
  fetchLogSnapshot,
  type SseStreamStatus,
} from "@/api/observability";

class FakeEventSource {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSED = 2;

  onopen: (() => void) | null = null;
  onmessage: ((e: MessageEvent) => void) | null = null;
  readyState = FakeEventSource.CONNECTING;
  private listeners: Record<string, ((e: Event) => void)[]> = {};

  close = vi.fn();

  addEventListener(type: string, cb: (e: Event) => void) {
    (this.listeners[type] ??= []).push(cb);
  }

  emit(type: string, event: Event) {
    this.listeners[type]?.forEach((cb) => cb(event));
  }
}

describe("attachSseHandlers", () => {
  let es: FakeEventSource;
  const onData = vi.fn();
  const onStreamError = vi.fn();
  const onStatusChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal("EventSource", FakeEventSource);
    es = new FakeEventSource();
    attachSseHandlers(es as unknown as EventSource, { onData, onStreamError, onStatusChange });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("reports connected on open and forwards message data", () => {
    es.onopen!();
    expect(onStatusChange).toHaveBeenCalledWith("connected");

    es.onmessage!(new MessageEvent("message", { data: "a log line" }));
    expect(onData).toHaveBeenCalledWith("a log line");
  });

  it("routes the backend's named error event to onStreamError and stops retrying", () => {
    es.emit("error", new MessageEvent("error", { data: "[web]: pod unreachable" }));
    expect(onStreamError).toHaveBeenCalledWith("[web]: pod unreachable");
    // Server closes after a terminal error; our side must too, or the browser
    // reconnect-loops and re-fires the error forever.
    expect(es.close).toHaveBeenCalled();
    expect(onStatusChange).toHaveBeenCalledWith("disconnected");
  });

  it("treats a native error while retrying as reconnecting, not an error", () => {
    es.readyState = FakeEventSource.CONNECTING;
    es.emit("error", new Event("error"));
    expect(onStatusChange).toHaveBeenCalledWith("reconnecting");
    expect(onStreamError).not.toHaveBeenCalled();
  });

  it("treats a native error after close as disconnected", () => {
    es.readyState = FakeEventSource.CLOSED;
    es.emit("error", new Event("error"));
    expect(onStatusChange).toHaveBeenCalledWith("disconnected");
    expect(onStreamError).not.toHaveBeenCalled();
  });
});

describe("aggregateStreamStatus", () => {
  const agg = (...statuses: SseStreamStatus[]) => aggregateStreamStatus(statuses);

  it("is connected while any stream delivers, even if another is dead", () => {
    expect(agg("connected", "connected")).toBe("connected");
    expect(agg("connected", "disconnected")).toBe("connected");
  });

  it("surfaces in-flight states over settled ones", () => {
    expect(agg("connected", "connecting")).toBe("connecting");
    expect(agg("connected", "reconnecting", "connecting")).toBe("reconnecting");
  });

  it("recovers to connected after a reconnect succeeds", () => {
    expect(agg("connected", "reconnecting")).toBe("reconnecting");
    expect(agg("connected", "connected")).toBe("connected");
  });

  it("is disconnected with no streams or all-dead streams", () => {
    expect(agg()).toBe("disconnected");
    expect(agg("disconnected", "disconnected")).toBe("disconnected");
  });
});

describe("fetchLogSnapshot", () => {
  const mockBody = (text: string, ok = true) =>
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok, text: () => Promise.resolve(text) }));

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns only the data payloads and drops the end frame", async () => {
    mockBody("data: line one\n\ndata: line two\n\nevent: end\ndata: {}\n\n");

    await expect(fetchLogSnapshot("o", "p", "s", "web")).resolves.toEqual(["line one", "line two"]);
  });

  it("drops a named error frame's payload too", async () => {
    mockBody("data: line one\n\nevent: error\ndata: [web]: pod unreachable\n\n");

    await expect(fetchLogSnapshot("o", "p", "s", "web")).resolves.toEqual(["line one"]);
  });

  it("keeps data lines that themselves look like SSE fields", async () => {
    mockBody("data: event: build started\n\ndata: data: nested\n\nevent: end\ndata: {}\n\n");

    await expect(fetchLogSnapshot("o", "p", "s", "web")).resolves.toEqual([
      "event: build started",
      "data: nested",
    ]);
  });

  it("returns [] on a failed response and on a thrown fetch", async () => {
    mockBody("data: line one\n\n", false);
    await expect(fetchLogSnapshot("o", "p", "s", "web")).resolves.toEqual([]);

    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
    await expect(fetchLogSnapshot("o", "p", "s", "web")).resolves.toEqual([]);
  });
});
