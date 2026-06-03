// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act, cleanup } from "@testing-library/react";

vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/teams", () => ({
  listTeams: vi.fn(), createTeam: vi.fn(), renameTeam: vi.fn(), deleteTeam: vi.fn(),
}));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, getErrorMessage: vi.fn(() => "nope") };
});

import { listTeams, createTeam, deleteTeam } from "@/api/teams";
import { useTeams } from "../use-teams";

beforeEach(() => { vi.mocked(listTeams).mockReset(); vi.mocked(createTeam).mockReset(); vi.mocked(deleteTeam).mockReset(); });
afterEach(() => cleanup());

describe("useTeams", () => {
  it("loads teams and computes onlyDefault", async () => {
    vi.mocked(listTeams).mockResolvedValue({ items: [{ id: "t1", name: "engineering", default_team: true }] } as never);
    const { result } = renderHook(() => useTeams());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.teams).toHaveLength(1);
    expect(result.current.onlyDefault).toBe(true);
  });

  it("create returns ok and refetches", async () => {
    vi.mocked(listTeams).mockResolvedValue({ items: [] } as never);
    vi.mocked(createTeam).mockResolvedValue({ id: "t2", name: "data" } as never);
    const { result } = renderHook(() => useTeams());
    await waitFor(() => expect(result.current.loading).toBe(false));
    let r: unknown;
    await act(async () => { r = await result.current.create("data"); });
    expect(createTeam).toHaveBeenCalledWith("org-1", { name: "data" });
    expect(r).toEqual({ ok: true });
  });

  it("remove maps failure", async () => {
    vi.mocked(listTeams).mockResolvedValue({ items: [] } as never);
    vi.mocked(deleteTeam).mockRejectedValue(new Error("x"));
    const { result } = renderHook(() => useTeams());
    await waitFor(() => expect(result.current.loading).toBe(false));
    let r: unknown;
    await act(async () => { r = await result.current.remove("data"); });
    expect(r).toEqual({ ok: false, error: "nope" });
  });
});
