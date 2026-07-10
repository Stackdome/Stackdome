// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";

vi.mock("@/api/preview-envs", async (importOriginal) => {
  const orig = await importOriginal<typeof import("@/api/preview-envs")>();
  return { ...orig, listPreviewEnvs: vi.fn() };
});
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-teams", () => ({
  useResourceTeams: () => ({ teams: [], teamNameById: () => undefined, defaultTeamName: "default" }),
}));

import { listPreviewEnvs } from "@/api/preview-envs";
import { usePreviewEnvs } from "../use-preview-envs";

beforeEach(() => vi.useFakeTimers());
afterEach(() => {
  vi.useRealTimers();
  vi.clearAllMocks();
});

describe("usePreviewEnvs", () => {
  it("loads envs and polls while a phase is non-terminal", async () => {
    (listPreviewEnvs as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({ items: [{ id: "p1", status: { phase: "Deploying" } }], total: 1 })
      .mockResolvedValueOnce({ items: [{ id: "p1", status: { phase: "Ready" } }], total: 1 });

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
    expect((listPreviewEnvs as ReturnType<typeof vi.fn>).mock.calls.length).toBe(2);
  });

  it("does not poll when everything is terminal", async () => {
    (listPreviewEnvs as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [{ id: "p1", status: { phase: "Ready" } }],
      total: 1,
    });
    renderHook(() => usePreviewEnvs("c1"));
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(30_000);
    });
    expect((listPreviewEnvs as ReturnType<typeof vi.fn>).mock.calls.length).toBe(1);
  });
});
