import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import { listProjects, createProject, renameProject, deleteProject, type Project } from "@/api/projects";

type ActionResult = { ok: true } | { ok: false; error: string };

export function useProjects() {
  const orgId = getCurrentOrganizationId();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProjects = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await listProjects(orgId);
      setProjects(res.items ?? []);
    } catch (e) {
      setError(getErrorMessage(e));
      setProjects([]);
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  const refetch = useCallback(() => { void fetchProjects(); }, [fetchProjects]);
  useEffect(() => { void fetchProjects(); }, [fetchProjects]);

  const create = useCallback(async (name: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await createProject(orgId, { name }); await fetchProjects(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, fetchProjects]);

  const rename = useCallback(async (projectName: string, newName: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await renameProject(orgId, projectName, { name: newName }); await fetchProjects(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, fetchProjects]);

  const remove = useCallback(async (projectName: string): Promise<ActionResult> => {
    if (!orgId) return { ok: false, error: "no organisation" };
    try { await deleteProject(orgId, projectName); await fetchProjects(); return { ok: true }; }
    catch (e) { return { ok: false, error: getErrorMessage(e) }; }
  }, [orgId, fetchProjects]);

  const onlyDefault = projects.length > 0 && projects.every((t) => t.default_project === true);

  return { projects, loading, error, refetch, create, rename, remove, onlyDefault };
}
