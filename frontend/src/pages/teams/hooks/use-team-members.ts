import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import {
  listTeamMembers, addTeamMember, updateTeamMemberRole, removeTeamMember,
  type TeamMembership,
} from "@/api/teams";

type ActionResult = { ok: true } | { ok: false; error: string };
type Role = "Developer" | "Viewer";

export function useTeamMembers(teamName: string) {
  const orgId = getCurrentOrganizationId();
  const [members, setMembers] = useState<TeamMembership[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchMembers = useCallback(async () => {
    if (!orgId || !teamName) return;
    setLoading(true);
    setError(null);
    try {
      const res = await listTeamMembers(orgId, teamName);
      setMembers(res.items ?? []);
    } catch (e) {
      setError(getErrorMessage(e));
      setMembers([]);
    } finally {
      setLoading(false);
    }
  }, [orgId, teamName]);

  const refetch = useCallback(() => { void fetchMembers(); }, [fetchMembers]);
  useEffect(() => { void fetchMembers(); }, [fetchMembers]);

  const addMember = useCallback(async (userId: string, role: Role): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await addTeamMember(orgId, teamName, { user_id: userId, role }); await fetchMembers(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, teamName, fetchMembers]);

  const changeRole = useCallback(async (membershipId: string, role: Role): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await updateTeamMemberRole(orgId, teamName, membershipId, { role }); await fetchMembers(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, teamName, fetchMembers]);

  const removeMember = useCallback(async (membershipId: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await removeTeamMember(orgId, teamName, membershipId); await fetchMembers(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, teamName, fetchMembers]);

  return { members, loading, error, refetch, addMember, changeRole, removeMember };
}
