import * as React from "react";
import type { User } from "@/api/users";
import { getCurrentUser as getStoredUser } from "@/helpers/common";
import { getCurrentUser as fetchCurrentUser } from "@/api/users";

type TeamRole = "Developer" | "Viewer";

interface CurrentUserValue {
  user: User | null;
  isOrgAdmin: boolean;
  organisationId: string | null;
  loading: boolean;
  refresh: () => Promise<void>;
  /** Role the current user holds in a team, matched by team_id or team_name. undefined if not a member. */
  roleInTeam: (teamRef: string) => TeamRole | undefined;
  /** Whether the current user may mutate resources in the given team (OrgAdmin anywhere, else Developer in that team). */
  canWrite: (teamRef: string) => boolean;
  /** Whether the user can write in at least one team (OrgAdmin, or Developer somewhere) — gates "create" entry points. */
  canWriteAnyTeam: boolean;
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

  const value = React.useMemo<CurrentUserValue>(() => {
    const isOrgAdmin = user?.role === "OrgAdmin";
    const roleInTeam = (teamRef: string): TeamRole | undefined =>
      user?.teams?.find((t) => t.team_id === teamRef || t.team_name === teamRef)?.role;
    const canWrite = (teamRef: string): boolean => isOrgAdmin || roleInTeam(teamRef) === "Developer";
    const canWriteAnyTeam = isOrgAdmin || (user?.teams?.some((t) => t.role === "Developer") ?? false);
    return {
      user,
      isOrgAdmin,
      organisationId: user?.organisation_id ?? null,
      loading,
      refresh,
      roleInTeam,
      canWrite,
      canWriteAnyTeam,
    };
  }, [user, loading, refresh]);
  return <CurrentUserContext.Provider value={value}>{children}</CurrentUserContext.Provider>;
}
