import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { listTeams, type Team } from "@/api/teams";

export function useTeamOptions() {
  const orgId = getCurrentOrganizationId();
  const [teams, setTeams] = useState<Team[]>([]);
  const [loading, setLoading] = useState(false);
  const fetchTeams = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    try {
      const res = await listTeams(orgId);
      setTeams(res.items ?? []);
    } catch {
      setTeams([]);
    } finally {
      setLoading(false);
    }
  }, [orgId]);
  useEffect(() => { void fetchTeams(); }, [fetchTeams]);
  return { teams, loading };
}
