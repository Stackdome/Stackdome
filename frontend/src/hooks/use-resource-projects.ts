import { useCallback, useEffect, useMemo, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { listProjects, type Project } from "@/api/projects";

/**
 * Resolves projects for resource writes (stacks/secrets/object-stores/addons),
 * which target project-scoped endpoints.
 *
 * Backed by the full org project list rather than the current user's memberships:
 * an OrgAdmin can write resources in projects they don't belong to, and new
 * resources default to the org's `default` project — neither is necessarily in
 * `user.projects`. (Role-based gating still uses `useCurrentUser().canWrite`.)
 */
export function useResourceProjects() {
  const orgId = getCurrentOrganizationId();
  const [projects, setProjects] = useState<Project[]>([]);

  useEffect(() => {
    if (!orgId) return;
    let cancelled = false;
    listProjects(orgId)
      .then((res) => {
        if (!cancelled) setProjects(res.items ?? []);
      })
      .catch(() => {
        // Non-fatal: callers guard on an unresolved project and surface an error.
      });
    return () => {
      cancelled = true;
    };
  }, [orgId]);

  const projectNameById = useCallback(
    (projectId: string | undefined): string | undefined =>
      projectId ? projects.find((t) => t.id === projectId)?.name : undefined,
    [projects],
  );

  const defaultProjectName = useMemo(() => projects.find((t) => t.default_project)?.name, [projects]);

  return { projects, projectNameById, defaultProjectName };
}
