import { useCallback, useEffect, useMemo, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { listTeams, type Team } from "@/api/teams";

/**
 * Resolves teams for resource writes (stacks/secrets/object-stores/addons),
 * which target team-scoped endpoints.
 *
 * Backed by the full org team list rather than the current user's memberships:
 * an OrgAdmin can write resources in teams they don't belong to, and new
 * resources default to the org's `default` team — neither is necessarily in
 * `user.teams`. (Role-based gating still uses `useCurrentUser().canWrite`.)
 */
export function useResourceTeams() {
  const orgId = getCurrentOrganizationId();
  const [teams, setTeams] = useState<Team[]>([]);

  useEffect(() => {
    if (!orgId) return;
    let cancelled = false;
    listTeams(orgId)
      .then((res) => {
        if (!cancelled) setTeams(res.items ?? []);
      })
      .catch(() => {
        // Non-fatal: callers guard on an unresolved team and surface an error.
      });
    return () => {
      cancelled = true;
    };
  }, [orgId]);

  const teamNameById = useCallback(
    (teamId: string | undefined): string | undefined =>
      teamId ? teams.find((t) => t.id === teamId)?.name : undefined,
    [teams],
  );

  const defaultTeamName = useMemo(() => teams.find((t) => t.default_team)?.name, [teams]);

  return { teams, teamNameById, defaultTeamName };
}
