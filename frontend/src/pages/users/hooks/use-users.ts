import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { listOrganizationUsers, type User } from "@/api/organizations";
import { listInvites, type OrgInvite } from "@/api/invites";

export interface ActiveRow {
  kind: "active";
  id: string;
  name: string;
  email: string;
  role: User["role"];
  teams: NonNullable<User["teams"]>;
  last_active_at?: string;
  user: User;
}
export interface PendingRow {
  kind: "pending";
  id: string;
  email: string;
  team_name?: string;
  role?: string;
  invited_by?: string;
  expires_at?: string;
  email_sent?: boolean;
  invite: OrgInvite;
}
export type UserRowModel = ActiveRow | PendingRow;

export function useUsers() {
  const orgId = getCurrentOrganizationId();
  const [rows, setRows] = useState<UserRowModel[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchAll = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    setError(null);
    try {
      const usersRes = await listOrganizationUsers(orgId);
      const active: ActiveRow[] = (usersRes.items ?? []).map((u) => ({
        kind: "active",
        id: u.id ?? "",
        name: u.name ?? u.email ?? "",
        email: u.email ?? "",
        role: u.role,
        teams: u.teams ?? [],
        last_active_at: undefined,
        user: u,
      }));
      let pending: PendingRow[] = [];
      try {
        const inv = await listInvites(orgId, "pending");
        pending = (inv.items ?? []).map((i) => ({
          kind: "pending",
          id: i.id ?? "",
          email: i.email ?? "",
          team_name: i.team_name,
          role: i.role,
          invited_by: i.invited_by,
          expires_at: i.expires_at,
          email_sent: i.email_sent,
          invite: i,
        }));
      } catch {
        pending = [];
      }
      setRows([...pending, ...active]);
    } catch (e: unknown) {
      setError(getErrorMessage(e));
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  const refetch = useCallback(() => { void fetchAll(); }, [fetchAll]);
  useEffect(() => { void fetchAll(); }, [fetchAll]);

  return { rows, loading, error, refetch };
}
