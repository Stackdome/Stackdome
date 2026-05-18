import * as React from "react";
import type { User } from "@/api/users";
import { getCurrentUser as getStoredUser } from "@/helpers/common";
import { getCurrentUser as fetchCurrentUser } from "@/api/users";

interface CurrentUserValue {
  user: User | null;
  isOrgAdmin: boolean;
  organisationId: string | null;
  loading: boolean;
  refresh: () => Promise<void>;
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

  const value: CurrentUserValue = {
    user,
    isOrgAdmin: user?.role === "OrgAdmin",
    organisationId: user?.organisation_id ?? null,
    loading,
    refresh,
  };
  return <CurrentUserContext.Provider value={value}>{children}</CurrentUserContext.Provider>;
}
