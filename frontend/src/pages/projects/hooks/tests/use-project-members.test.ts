// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act, cleanup } from "@testing-library/react";

vi.mock("@/lib/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/projects", () => ({
  listProjectMembers: vi.fn(), addProjectMember: vi.fn(), updateProjectMemberRole: vi.fn(), removeProjectMember: vi.fn(),
}));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, getErrorMessage: vi.fn(() => "err") };
});

import { listProjectMembers, addProjectMember, updateProjectMemberRole, removeProjectMember } from "@/api/projects";
import { useProjectMembers } from "../use-project-members";

beforeEach(() => {
  vi.mocked(listProjectMembers).mockReset(); vi.mocked(addProjectMember).mockReset();
  vi.mocked(updateProjectMemberRole).mockReset(); vi.mocked(removeProjectMember).mockReset();
});
afterEach(() => cleanup());

describe("useProjectMembers", () => {
  it("loads members for the project", async () => {
    vi.mocked(listProjectMembers).mockResolvedValue({ items: [{ id: "m1", role: "Developer" }] } as never);
    const { result } = renderHook(() => useProjectMembers("engineering"));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.members).toHaveLength(1);
    expect(listProjectMembers).toHaveBeenCalledWith("org-1", "engineering");
  });

  it("addMember posts user_id+role and refetches", async () => {
    vi.mocked(listProjectMembers).mockResolvedValue({ items: [] } as never);
    vi.mocked(addProjectMember).mockResolvedValue({ id: "m2" } as never);
    const { result } = renderHook(() => useProjectMembers("engineering"));
    await waitFor(() => expect(result.current.loading).toBe(false));
    let r: unknown;
    await act(async () => { r = await result.current.addMember("u9", "Viewer"); });
    expect(addProjectMember).toHaveBeenCalledWith("org-1", "engineering", { user_id: "u9", role: "Viewer" });
    expect(r).toEqual({ ok: true });
  });

  it("changeRole and removeMember map failures", async () => {
    vi.mocked(listProjectMembers).mockResolvedValue({ items: [] } as never);
    vi.mocked(updateProjectMemberRole).mockRejectedValue(new Error("x"));
    vi.mocked(removeProjectMember).mockRejectedValue(new Error("y"));
    const { result } = renderHook(() => useProjectMembers("engineering"));
    await waitFor(() => expect(result.current.loading).toBe(false));
    let a: unknown, b: unknown;
    await act(async () => { a = await result.current.changeRole("m1", "Viewer"); b = await result.current.removeMember("m1"); });
    expect(a).toEqual({ ok: false, error: "err" });
    expect(b).toEqual({ ok: false, error: "err" });
  });
});
