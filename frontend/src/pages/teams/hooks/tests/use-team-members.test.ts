// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act, cleanup } from "@testing-library/react";

vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/teams", () => ({
  listTeamMembers: vi.fn(), addTeamMember: vi.fn(), updateTeamMemberRole: vi.fn(), removeTeamMember: vi.fn(),
}));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, getErrorMessage: vi.fn(() => "err") };
});

import { listTeamMembers, addTeamMember, updateTeamMemberRole, removeTeamMember } from "@/api/teams";
import { useTeamMembers } from "../use-team-members";

beforeEach(() => {
  vi.mocked(listTeamMembers).mockReset(); vi.mocked(addTeamMember).mockReset();
  vi.mocked(updateTeamMemberRole).mockReset(); vi.mocked(removeTeamMember).mockReset();
});
afterEach(() => cleanup());

describe("useTeamMembers", () => {
  it("loads members for the team", async () => {
    vi.mocked(listTeamMembers).mockResolvedValue({ items: [{ id: "m1", role: "Developer" }] } as never);
    const { result } = renderHook(() => useTeamMembers("engineering"));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.members).toHaveLength(1);
    expect(listTeamMembers).toHaveBeenCalledWith("org-1", "engineering");
  });

  it("addMember posts user_id+role and refetches", async () => {
    vi.mocked(listTeamMembers).mockResolvedValue({ items: [] } as never);
    vi.mocked(addTeamMember).mockResolvedValue({ id: "m2" } as never);
    const { result } = renderHook(() => useTeamMembers("engineering"));
    await waitFor(() => expect(result.current.loading).toBe(false));
    let r: unknown;
    await act(async () => { r = await result.current.addMember("u9", "Viewer"); });
    expect(addTeamMember).toHaveBeenCalledWith("org-1", "engineering", { user_id: "u9", role: "Viewer" });
    expect(r).toEqual({ ok: true });
  });

  it("changeRole and removeMember map failures", async () => {
    vi.mocked(listTeamMembers).mockResolvedValue({ items: [] } as never);
    vi.mocked(updateTeamMemberRole).mockRejectedValue(new Error("x"));
    vi.mocked(removeTeamMember).mockRejectedValue(new Error("y"));
    const { result } = renderHook(() => useTeamMembers("engineering"));
    await waitFor(() => expect(result.current.loading).toBe(false));
    let a: unknown, b: unknown;
    await act(async () => { a = await result.current.changeRole("m1", "Viewer"); b = await result.current.removeMember("m1"); });
    expect(a).toEqual({ ok: false, error: "err" });
    expect(b).toEqual({ ok: false, error: "err" });
  });
});
