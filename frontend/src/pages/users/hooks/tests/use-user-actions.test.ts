// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, cleanup } from "@testing-library/react";

vi.mock("@/lib/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/organizations", () => ({ promoteAdmin: vi.fn(), demoteAdmin: vi.fn() }));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, getErrorMessage: vi.fn(() => "denied") };
});

import { promoteAdmin, demoteAdmin } from "@/api/organizations";
import { useUserActions } from "../use-user-actions";

beforeEach(() => { vi.mocked(promoteAdmin).mockReset(); vi.mocked(demoteAdmin).mockReset(); });
afterEach(() => cleanup());

describe("useUserActions", () => {
  it("promote returns ok on success", async () => {
    vi.mocked(promoteAdmin).mockResolvedValue(undefined as never);
    const { result } = renderHook(() => useUserActions());
    let r: unknown;
    await act(async () => { r = await result.current.promote("u1"); });
    expect(promoteAdmin).toHaveBeenCalledWith("org-1", { user_id: "u1" });
    expect(r).toEqual({ ok: true });
  });

  it("demote maps error to { ok:false, error }", async () => {
    vi.mocked(demoteAdmin).mockRejectedValue(new Error("x"));
    const { result } = renderHook(() => useUserActions());
    let r: unknown;
    await act(async () => { r = await result.current.demote("u1", "engineering", "Viewer"); });
    expect(demoteAdmin).toHaveBeenCalledWith("org-1", "u1", { project_name: "engineering", role: "Viewer" });
    expect(r).toEqual({ ok: false, error: "denied" });
  });
});
