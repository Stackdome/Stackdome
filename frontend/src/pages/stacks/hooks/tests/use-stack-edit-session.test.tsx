// @vitest-environment jsdom
import { renderHook, act } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { useStackEditSession } from "../use-stack-edit-session";

describe("rebase", () => {
  it("advances the baseline without touching the draft", () => {
    const { result } = renderHook(() => useStackEditSession());
    act(() => result.current.start({ resources: [{ name: "web" }], volumes: [] }));
    act(() => result.current.updateResources((prev) => [{ ...prev[0], name: "web2" }]));
    expect(result.current.dirty.dirtyResourceIdx.size).toBe(1);

    const snapshot = { resources: result.current.draft.resources, volumes: result.current.draft.volumes };
    act(() => result.current.rebase(snapshot));
    expect(result.current.dirty.dirtyResourceIdx.size).toBe(0);
    expect(result.current.draft.resources[0].name).toBe("web2");
  });

  it("is a no-op when the session is inactive", () => {
    const { result } = renderHook(() => useStackEditSession());
    act(() => result.current.rebase({ resources: [{ name: "x" }], volumes: [] }));
    expect(result.current.isActive).toBe(false);
    expect(result.current.baseline.resources).toEqual([]);
  });
});

describe("start with a split draft", () => {
  it("seeds draft and baseline independently so pre-existing autosaved edits open dirty", () => {
    const { result } = renderHook(() => useStackEditSession());
    // Baseline = deployed snapshot; draft = server state that already diverged.
    act(() =>
      result.current.start(
        { resources: [{ name: "web", source: { image: { ref: "nginx:1" } } }], volumes: [] },
        { draft: { resources: [{ name: "web", source: { image: { ref: "nginx:2" } } }], volumes: [] } },
      ));
    expect(result.current.draft.resources[0].source?.image?.ref).toBe("nginx:2");
    expect(result.current.baseline.resources[0].source?.image?.ref).toBe("nginx:1");
    expect(result.current.dirty.dirtyResourceIdx.size).toBe(1);

    // Per-resource discard restores the deployed value, not the autosaved one.
    act(() => result.current.discardResource(0));
    expect(result.current.draft.resources[0].source?.image?.ref).toBe("nginx:1");
    expect(result.current.dirty.dirtyResourceIdx.size).toBe(0);
  });
});
