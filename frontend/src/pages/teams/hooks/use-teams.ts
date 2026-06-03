import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { listTeams, createTeam, renameTeam, deleteTeam, type Team } from "@/api/teams";

type ActionResult = { ok: true } | { ok: false; error: string };

export function useTeams() {
  const orgId = getCurrentOrganizationId();
  const [teams, setTeams] = useState<Team[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchTeams = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await listTeams(orgId);
      setTeams(res.items ?? []);
    } catch (e) {
      setError(getErrorMessage(e));
      setTeams([]);
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  const refetch = useCallback(() => { void fetchTeams(); }, [fetchTeams]);
  useEffect(() => { void fetchTeams(); }, [fetchTeams]);

  const create = useCallback(async (name: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await createTeam(orgId, { name }); await fetchTeams(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, fetchTeams]);

  const rename = useCallback(async (teamName: string, newName: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await renameTeam(orgId, teamName, { name: newName }); await fetchTeams(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, fetchTeams]);

  const remove = useCallback(async (teamName: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await deleteTeam(orgId, teamName); await fetchTeams(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, fetchTeams]);

  const onlyDefault = teams.length > 0 && teams.every((t) => t.default_team === true);

  return { teams, loading, error, refetch, create, rename, remove, onlyDefault };
}
