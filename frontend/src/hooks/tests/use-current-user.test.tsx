// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, cleanup } from "@testing-library/react";
import React from "react";

vi.mock("@/helpers/common", () => ({
  getCurrentUser: vi.fn(),
}));
vi.mock("@/api/users", () => ({ getCurrentUser: vi.fn() }));

import { getCurrentUser as getStoredUser } from "@/helpers/common";
import { getCurrentUser as fetchCurrentUser } from "@/api/users";
import { CurrentUserProvider } from "@/contexts/current-user-context";
import { useCurrentUser } from "@/hooks/use-current-user";

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <CurrentUserProvider>{children}</CurrentUserProvider>
);

beforeEach(() => {
  vi.mocked(getStoredUser).mockReset();
  vi.mocked(fetchCurrentUser).mockReset();
});
afterEach(() => cleanup());

describe("useCurrentUser", () => {
  it("derives isOrgAdmin=true from a stored OrgAdmin user immediately", () => {
    vi.mocked(getStoredUser).mockReturnValue({ id: "u1", role: "OrgAdmin", organisation_id: "org-1" } as never);
    vi.mocked(fetchCurrentUser).mockResolvedValue({ id: "u1", role: "OrgAdmin", organisation_id: "org-1" } as never);
    const { result } = renderHook(() => useCurrentUser(), { wrapper });
    expect(result.current.isOrgAdmin).toBe(true);
    expect(result.current.organisationId).toBe("org-1");
  });

  it("isOrgAdmin=false for OrgMember", () => {
    vi.mocked(getStoredUser).mockReturnValue({ id: "u2", role: "OrgMember", organisation_id: "org-1" } as never);
    vi.mocked(fetchCurrentUser).mockResolvedValue({ id: "u2", role: "OrgMember", organisation_id: "org-1" } as never);
    const { result } = renderHook(() => useCurrentUser(), { wrapper });
    expect(result.current.isOrgAdmin).toBe(false);
  });

  it("refreshes from the API and updates role", async () => {
    vi.mocked(getStoredUser).mockReturnValue({ id: "u3", role: "OrgMember", organisation_id: "org-1" } as never);
    vi.mocked(fetchCurrentUser).mockResolvedValue({ id: "u3", role: "OrgAdmin", organisation_id: "org-1" } as never);
    const { result } = renderHook(() => useCurrentUser(), { wrapper });
    await waitFor(() => expect(result.current.isOrgAdmin).toBe(true));
  });

  it("retains the stored user when the refresh API call fails", async () => {
    vi.mocked(getStoredUser).mockReturnValue({ id: "u4", role: "OrgAdmin", organisation_id: "org-1" } as never);
    vi.mocked(fetchCurrentUser).mockRejectedValue(new Error("network down"));
    const { result } = renderHook(() => useCurrentUser(), { wrapper });
    expect(result.current.isOrgAdmin).toBe(true);
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.isOrgAdmin).toBe(true);
    expect(result.current.organisationId).toBe("org-1");
  });
});
