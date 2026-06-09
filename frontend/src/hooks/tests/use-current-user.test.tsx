// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, cleanup, act } from "@testing-library/react";
import React from "react";
import { AUTH_SESSION_CHANGED } from "@/helpers/auth-events";

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

  it("re-hydrates on AUTH_SESSION_CHANGED (post-login/signup without a reload)", async () => {
    // Initial mount mimics landing on /sign-in: no stored user, fetch unauthenticated.
    vi.mocked(getStoredUser).mockReturnValue(null as never);
    vi.mocked(fetchCurrentUser).mockRejectedValue(new Error("401"));
    const { result } = renderHook(() => useCurrentUser(), { wrapper });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.isOrgAdmin).toBe(false);
    expect(result.current.organisationId).toBeNull();

    // User logs in: session is now stored and the token is valid.
    vi.mocked(getStoredUser).mockReturnValue({ id: "u5", role: "OrgAdmin", organisation_id: "org-1" } as never);
    vi.mocked(fetchCurrentUser).mockResolvedValue({ id: "u5", role: "OrgAdmin", organisation_id: "org-1" } as never);
    act(() => {
      window.dispatchEvent(new Event(AUTH_SESSION_CHANGED));
    });

    await waitFor(() => expect(result.current.isOrgAdmin).toBe(true));
    expect(result.current.organisationId).toBe("org-1");
  });
});

describe("useCurrentUser team-role helpers", () => {
  const memberUser = {
    id: "m1",
    role: "OrgMember",
    organisation_id: "org-1",
    teams: [
      { team_id: "t1", team_name: "alpha", role: "Developer", default_team: true },
      { team_id: "t2", team_name: "beta", role: "Viewer", default_team: false },
    ],
  };
  const adminUser = {
    id: "a1",
    role: "OrgAdmin",
    organisation_id: "org-1",
    teams: [{ team_id: "t1", team_name: "alpha", role: "Viewer", default_team: true }],
  };

  function mockUser(u: unknown) {
    vi.mocked(getStoredUser).mockReturnValue(u as never);
    vi.mocked(fetchCurrentUser).mockResolvedValue(u as never);
  }

  describe("roleInTeam", () => {
    it("resolves the role by team_id", () => {
      mockUser(memberUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.roleInTeam("t1")).toBe("Developer");
      expect(result.current.roleInTeam("t2")).toBe("Viewer");
    });

    it("resolves the role by team_name", () => {
      mockUser(memberUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.roleInTeam("alpha")).toBe("Developer");
      expect(result.current.roleInTeam("beta")).toBe("Viewer");
    });

    it("returns undefined for a team the user does not belong to", () => {
      mockUser(memberUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.roleInTeam("t99")).toBeUndefined();
      expect(result.current.roleInTeam("gamma")).toBeUndefined();
    });

    it("returns undefined when there is no user", () => {
      mockUser(null);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.roleInTeam("t1")).toBeUndefined();
    });
  });

  describe("canWrite", () => {
    it("is true for an OrgAdmin regardless of team (even a Viewer membership or unknown team)", () => {
      mockUser(adminUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWrite("t1")).toBe(true);
      expect(result.current.canWrite("t99")).toBe(true);
    });

    it("is true for a Developer in the team", () => {
      mockUser(memberUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWrite("t1")).toBe(true);
      expect(result.current.canWrite("alpha")).toBe(true);
    });

    it("is false for a Viewer in the team", () => {
      mockUser(memberUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWrite("t2")).toBe(false);
      expect(result.current.canWrite("beta")).toBe(false);
    });

    it("is false for a non-member team", () => {
      mockUser(memberUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWrite("t99")).toBe(false);
    });

    it("handles a multi-team user: Developer in A, Viewer in B", () => {
      mockUser(memberUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWrite("t1")).toBe(true);
      expect(result.current.canWrite("t2")).toBe(false);
    });
  });

  describe("canWriteAnyTeam", () => {
    it("is true for an OrgAdmin", () => {
      mockUser(adminUser);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWriteAnyTeam).toBe(true);
    });

    it("is true for a member who is Developer in at least one team", () => {
      mockUser(memberUser); // Developer in alpha, Viewer in beta
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWriteAnyTeam).toBe(true);
    });

    it("is false for a member who is only ever a Viewer", () => {
      mockUser({
        id: "v1",
        role: "OrgMember",
        organisation_id: "org-1",
        teams: [
          { team_id: "t2", team_name: "beta", role: "Viewer", default_team: true },
          { team_id: "t3", team_name: "gamma", role: "Viewer", default_team: false },
        ],
      });
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWriteAnyTeam).toBe(false);
    });

    it("is false when there is no user", () => {
      mockUser(null);
      const { result } = renderHook(() => useCurrentUser(), { wrapper });
      expect(result.current.canWriteAnyTeam).toBe(false);
    });
  });
});
