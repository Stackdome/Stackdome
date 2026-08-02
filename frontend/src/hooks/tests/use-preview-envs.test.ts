// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";

vi.mock("@/api/preview-envs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-envs")>();
  return { ...orig, listAllPreviewEnvs: vi.fn() };
});
vi.mock("@/lib/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({ projects: [], projectNameById: () => undefined, defaultProjectName: "default" }),
}));

import { listAllPreviewEnvs } from "@/api/preview-envs";
import { usePreviewEnvs } from "@/hooks/use-preview-envs";

beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  vi.useRealTimers();
  vi.clearAllMocks();
});

describe("usePreviewEnvs", () => {
  it("loads envs and polls while a phase is non-terminal", async () => {
    (listAllPreviewEnvs as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce([{ id: "p1", status: { phase: "Deploying" } }])
      .mockResolvedValueOnce([{ id: "p1", status: { phase: "Ready" } }]);

    const { result } = renderHook(() => usePreviewEnvs("c1"));
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    expect(result.current.envs[0]?.status?.phase).toBe("Deploying");

    await act(async () => {
      await vi.advanceTimersByTimeAsync(7_000);
    });
    expect(result.current.envs[0]?.status?.phase).toBe("Ready");

    // All terminal → no further polls.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(21_000);
    });
    expect((listAllPreviewEnvs as ReturnType<typeof vi.fn>).mock.calls.length).toBe(2);
  });

  it("does not poll when everything is terminal", async () => {
    (listAllPreviewEnvs as ReturnType<typeof vi.fn>).mockResolvedValue([{ id: "p1", status: { phase: "Ready" } }]);
    renderHook(() => usePreviewEnvs("c1"));
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(30_000);
    });
    expect((listAllPreviewEnvs as ReturnType<typeof vi.fn>).mock.calls.length).toBe(1);
  });

  it("keeps polling when an env has no status yet (phase not reported)", async () => {
    (listAllPreviewEnvs as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce([{ id: "p1" }])
      .mockResolvedValueOnce([{ id: "p1" }]);

    renderHook(() => usePreviewEnvs("c1"));
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    expect((listAllPreviewEnvs as ReturnType<typeof vi.fn>).mock.calls.length).toBe(1);

    // Second interval tick should still fire because a missing phase counts
    // as non-terminal (keep polling until a terminal phase appears).
    await act(async () => {
      await vi.advanceTimersByTimeAsync(7_000);
    });
    expect((listAllPreviewEnvs as ReturnType<typeof vi.fn>).mock.calls.length).toBe(2);
  });
});
