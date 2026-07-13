// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, cleanup } from "@testing-library/react";

vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: vi.fn(() => "org-1") }));
vi.mock("@/api/organizations", () => ({ listOrganizationUsers: vi.fn() }));
vi.mock("@/api/invites", () => ({ listInvites: vi.fn() }));
vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return { ...actual, getErrorMessage: vi.fn(() => "boom") };
});

import { listOrganizationUsers } from "@/api/organizations";
import { listInvites } from "@/api/invites";
import { useUsers } from "../use-users";

const mockedUsers = vi.mocked(listOrganizationUsers);
const mockedInvites = vi.mocked(listInvites);

beforeEach(() => { mockedUsers.mockReset(); mockedInvites.mockReset(); });
afterEach(() => cleanup());

describe("useUsers", () => {
  it("merges pending invites above active users", async () => {
    mockedUsers.mockResolvedValue({ items: [{ id: "u1", name: "Ada", email: "ada@x.io", role: "OrgMember" }] } as never);
    mockedInvites.mockResolvedValue({ items: [{ id: "inv1", email: "neo@x.io", status: "pending", project_name: "engineering", role: "Developer", expires_at: "2026-05-19T00:00:00Z", invited_by: "ada", email_sent: true }] } as never);
    const { result } = renderHook(() => useUsers());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.rows.map(r => r.kind)).toEqual(["pending", "active"]);
    expect(result.current.rows[0]).toMatchObject({ kind: "pending", email: "neo@x.io" });
    expect(result.current.rows[1]).toMatchObject({ kind: "active", email: "ada@x.io" });
  });

  it("still shows users when the invites call fails", async () => {
    mockedUsers.mockResolvedValue({ items: [{ id: "u1", name: "Ada", email: "ada@x.io", role: "OrgMember" }] } as never);
    mockedInvites.mockRejectedValue(new Error("invites down"));
    const { result } = renderHook(() => useUsers());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.rows).toHaveLength(1);
    expect(result.current.rows[0].kind).toBe("active");
    expect(result.current.error).toBeNull();
  });

  it("sets error when the users call fails", async () => {
    mockedUsers.mockRejectedValue(new Error("users down"));
    mockedInvites.mockResolvedValue({ items: [] } as never);
    const { result } = renderHook(() => useUsers());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe("boom");
    expect(result.current.rows).toHaveLength(0);
  });
});
