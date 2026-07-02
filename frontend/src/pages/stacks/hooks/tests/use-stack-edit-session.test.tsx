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
