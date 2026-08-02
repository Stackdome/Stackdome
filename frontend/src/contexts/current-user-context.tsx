import * as React from "react";
import type { User } from "@/api/users";
import { getCurrentUser as getStoredUser } from "@/lib/common";
import { getCurrentUser as fetchCurrentUser } from "@/api/users";
import { AUTH_SESSION_CHANGED } from "@/lib/auth-events";

type ProjectRole = "Developer" | "Viewer";

interface CurrentUserValue {
  user: User | null;
  isOrgAdmin: boolean;
  organisationId: string | null;
  loading: boolean;
  refresh: () => Promise<void>;
  /** Role the current user holds in a project, matched by project_id or project_name. undefined if not a member. */
  roleInProject: (projectRef: string) => ProjectRole | undefined;
  /** Whether the current user may mutate resources in the given project (OrgAdmin anywhere, else Developer in that project). */
  canWrite: (projectRef: string) => boolean;
  /** Whether the user can write in at least one project (OrgAdmin, or Developer somewhere) — gates "create" entry points. */
  canWriteAnyProject: boolean;
}

export const CurrentUserContext = React.createContext<CurrentUserValue | undefined>(undefined);

export function CurrentUserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = React.useState<User | null>(() => getStoredUser());
  const [loading, setLoading] = React.useState(false);

  const refresh = React.useCallback(async () => {
    setLoading(true);
    try {
      const fresh = await fetchCurrentUser();
      setUser(fresh);
    } catch {
      // keep last-known user; gate falls back to stored role
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    void refresh();
  }, [refresh]);

  // Re-hydrate after login/signup/logout. The provider mounts once (on /sign-in,
  // before a token exists), and client-side navigation after auth does not remount
  // it — so without this the gate stays stuck on the pre-auth (null) user.
  React.useEffect(() => {
    const onAuthChange = () => {
      setUser(getStoredUser());
      void refresh();
    };
    window.addEventListener(AUTH_SESSION_CHANGED, onAuthChange);
    return () => window.removeEventListener(AUTH_SESSION_CHANGED, onAuthChange);
  }, [refresh]);

  const value = React.useMemo<CurrentUserValue>(() => {
    const isOrgAdmin = user?.role === "OrgAdmin";
    const roleInProject = (projectRef: string): ProjectRole | undefined =>
      user?.projects?.find((t) => t.project_id === projectRef || t.project_name === projectRef)?.role;
    const canWrite = (projectRef: string): boolean => isOrgAdmin || roleInProject(projectRef) === "Developer";
    const canWriteAnyProject = isOrgAdmin || (user?.projects?.some((t) => t.role === "Developer") ?? false);
    return {
      user,
      isOrgAdmin,
      organisationId: user?.organisation_id ?? null,
      loading,
      refresh,
      roleInProject,
      canWrite,
      canWriteAnyProject,
    };
  }, [user, loading, refresh]);
  return <CurrentUserContext.Provider value={value}>{children}</CurrentUserContext.Provider>;
}
