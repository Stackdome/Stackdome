import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import { listOrganizationUsers, type User } from "@/api/organizations";
import { listInvites, type OrgInvite } from "@/api/invites";

export interface ActiveRow {
  kind: "active";
  id: string;
  name: string;
  email: string;
  role: User["role"];
  projects: NonNullable<User["projects"]>;
  user: User;
}
export interface PendingRow {
  kind: "pending";
  id: string;
  email: string;
  project_name?: string;
  role?: OrgInvite["role"];
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
        projects: u.projects ?? [],
        user: u,
      }));
      let pending: PendingRow[] = [];
      try {
        const inv = await listInvites(orgId, "pending");
        pending = (inv.items ?? []).map((i) => ({
          kind: "pending",
          id: i.id ?? "",
          email: i.email ?? "",
          project_name: i.project_name,
          role: i.role,
          invited_by: i.invited_by,
          expires_at: i.expires_at,
          email_sent: i.email_sent,
          invite: i,
        }));
      } catch (e) {
        console.warn("Failed to fetch pending invites (non-fatal):", e);
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
