import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import {
  listProjectMembers, addProjectMember, updateProjectMemberRole, removeProjectMember,
  type ProjectMembership,
} from "@/api/projects";

type ActionResult = { ok: true } | { ok: false; error: string };
type Role = "Developer" | "Viewer";

export function useProjectMembers(projectName: string) {
  const orgId = getCurrentOrganizationId();
  const [members, setMembers] = useState<ProjectMembership[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchMembers = useCallback(async () => {
    if (!orgId || !projectName) return;
    setLoading(true);
    setError(null);
    try {
      const res = await listProjectMembers(orgId, projectName);
      setMembers(res.items ?? []);
    } catch (e) {
      setError(getErrorMessage(e));
      setMembers([]);
    } finally {
      setLoading(false);
    }
  }, [orgId, projectName]);

  const refetch = useCallback(() => { void fetchMembers(); }, [fetchMembers]);
  useEffect(() => { void fetchMembers(); }, [fetchMembers]);

  const addMember = useCallback(async (userId: string, role: Role): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await addProjectMember(orgId, projectName, { user_id: userId, role }); await fetchMembers(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, projectName, fetchMembers]);

  const changeRole = useCallback(async (membershipId: string, role: Role): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await updateProjectMemberRole(orgId, projectName, membershipId, { role }); await fetchMembers(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, projectName, fetchMembers]);

  const removeMember = useCallback(async (membershipId: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await removeProjectMember(orgId, projectName, membershipId); await fetchMembers(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, projectName, fetchMembers]);

  return { members, loading, error, refetch, addMember, changeRole, removeMember };
}
